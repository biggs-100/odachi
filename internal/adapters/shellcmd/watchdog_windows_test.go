//go:build windows

package shellcmd_test

func quietCommand() string { return "ping -n 3 127.0.0.1 >nul" }

func chattyCommand() string {
	// cmd's echo is the first byte, before PowerShell has finished starting;
	// the timed lines after it are what re-arm the idle clock.
	return `echo a & powershell -NoProfile -Command "Start-Sleep -Milliseconds 150; Write-Output b; Start-Sleep -Milliseconds 150; Write-Output c; Start-Sleep -Milliseconds 150; Write-Output d"`
}

func quietAfterFirstCommand() string { return "echo first & ping -n 3 127.0.0.1 >nul" }
