//go:build windows
// +build windows

package elevate

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modshell32          = windows.NewLazySystemDLL("shell32.dll")
	procShellExecuteExW = modshell32.NewProc("ShellExecuteExW")
)

const (
	seeMaskNoCloseProcess = 0x00000040
	seeMaskNoAsync        = 0x00000100
	swHide                = 0

	errorCancelled = 1223 // ERROR_CANCELLED — user declined the UAC prompt
)

// shellExecuteInfo mirrors the Win32 SHELLEXECUTEINFOW struct exactly.
// Field order and types must match or ShellExecuteEx will misread it.
type shellExecuteInfo struct {
	cbSize         uint32
	fMask          uint32
	hwnd           windows.Handle
	lpVerb         *uint16
	lpFile         *uint16
	lpParameters   *uint16
	lpDirectory    *uint16
	nShow          int32
	hInstApp       windows.Handle
	lpIDList       uintptr
	lpClass        *uint16
	hkeyClass      windows.Handle
	dwHotKey       uint32
	hIconOrMonitor windows.Handle
	hProcess       windows.Handle
}

// runElevated launches the command through cmd.exe with the "runas" verb so
// Windows shows the UAC consent dialog. Because ShellExecuteEx cannot pipe
// stdout/stderr, we redirect them to temp files inside the wrapped command
// line and read those back after the elevated process exits.
func runElevated(ctx context.Context, r Request) (Result, error) {
	outPath, errPath, cleanup, err := makeTempPair()
	if err != nil {
		return Result{ExitCode: -1}, err
	}
	defer cleanup()

	// Build:  /c <command> <args> > "out" 2> "err"
	inner := buildCmdLine(r.Command, r.Args)
	params := fmt.Sprintf(`/c %s > "%s" 2> "%s"`, inner, outPath, errPath)

	verbPtr, _ := windows.UTF16PtrFromString("runas")
	filePtr, _ := windows.UTF16PtrFromString("cmd.exe")
	paramsPtr, _ := windows.UTF16PtrFromString(params)

	info := shellExecuteInfo{
		fMask:        seeMaskNoCloseProcess | seeMaskNoAsync,
		lpVerb:       verbPtr,
		lpFile:       filePtr,
		lpParameters: paramsPtr,
		nShow:        swHide,
	}
	info.cbSize = uint32(unsafe.Sizeof(info))

	ret, _, callErr := procShellExecuteExW.Call(uintptr(unsafe.Pointer(&info)))
	if ret == 0 {
		// ShellExecuteEx failed. The most common cause is the user clicking
		// "No" on the UAC prompt → ERROR_CANCELLED.
		if errno, ok := callErr.(windows.Errno); ok && uint32(errno) == errorCancelled {
			return Result{ExitCode: -1, Cancelled: true}, nil
		}
		// Some Windows builds surface the cancel via GetLastError instead.
		if errno, ok := windows.GetLastError().(windows.Errno); ok && uint32(errno) == errorCancelled {
			return Result{ExitCode: -1, Cancelled: true}, nil
		}
		return Result{ExitCode: -1}, fmt.Errorf("ShellExecuteEx falló: %v", callErr)
	}
	if info.hProcess == 0 {
		return Result{ExitCode: -1}, fmt.Errorf("no se obtuvo handle del proceso elevado")
	}
	defer windows.CloseHandle(info.hProcess)

	if err := waitProcess(ctx, info.hProcess); err != nil {
		return Result{ExitCode: -1}, err
	}

	var code uint32
	_ = windows.GetExitCodeProcess(info.hProcess, &code)

	stdout, _ := os.ReadFile(outPath)
	stderr, _ := os.ReadFile(errPath)

	return Result{
		OK:       code == 0,
		ExitCode: int(code),
		Stdout:   decodeConsole(stdout),
		Stderr:   decodeConsole(stderr),
	}, nil
}

// waitProcess blocks until the elevated process exits or ctx is cancelled.
func waitProcess(ctx context.Context, h windows.Handle) error {
	for {
		// Wait in 250 ms slices so we can honour ctx cancellation.
		ev, err := windows.WaitForSingleObject(h, 250)
		if err != nil {
			return fmt.Errorf("WaitForSingleObject: %w", err)
		}
		if ev == windows.WAIT_OBJECT_0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
}

func makeTempPair() (outPath, errPath string, cleanup func(), err error) {
	of, err := os.CreateTemp("", "vop-elev-out-*.txt")
	if err != nil {
		return "", "", func() {}, err
	}
	ef, err := os.CreateTemp("", "vop-elev-err-*.txt")
	if err != nil {
		of.Close()
		os.Remove(of.Name())
		return "", "", func() {}, err
	}
	of.Close()
	ef.Close()
	return of.Name(), ef.Name(), func() {
		os.Remove(of.Name())
		os.Remove(ef.Name())
	}, nil
}

// buildCmdLine renders the command + args for cmd.exe, quoting tokens that
// contain spaces or quotes.
func buildCmdLine(command string, args []string) string {
	parts := append([]string{command}, args...)
	out := make([]string, len(parts))
	for i, p := range parts {
		if strings.ContainsAny(p, " \t\"") {
			out[i] = `"` + strings.ReplaceAll(p, `"`, `""`) + `"`
		} else {
			out[i] = p
		}
	}
	return strings.Join(out, " ")
}

// decodeConsole returns the bytes as a string. Windows console tools emit
// OEM/UTF-8 mixed output; for the educational tools we run (winget, vboxmanage,
// vagrant) UTF-8 is the common case, so we pass through and trim.
func decodeConsole(b []byte) string {
	return strings.TrimRight(string(b), "\r\n ")
}

var _ = time.Second // reserved for future timeout tuning
