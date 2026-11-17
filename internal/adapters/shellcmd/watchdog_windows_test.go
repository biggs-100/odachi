//go:build windows

package shellcmd_test

func quietCommand() string { return "ping -n 3 127.0.0.1 >nul" }

func chattyCommand() string {
	// cmd's echo is the first byte, before PowerShell has finished starting.
	// Ten lines at 300 ms each outlive the idle clock's two seconds only if the
	// clock is re-armed; a slow CI runner starting PowerShell is not a stall.
	return `echo a & powershell -NoProfile -Command "1..10 | ForEach-Object { Start-Sleep -Milliseconds 300; Write-Output $_ }"`
}

func quietAfterFirstCommand() string { return "echo first & ping -n 3 127.0.0.1 >nul" }
