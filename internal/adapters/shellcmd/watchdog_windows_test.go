//go:build windows

package shellcmd_test

func quietCommand() string { return "ping -n 3 127.0.0.1 >nul" }

func chattyCommand() string {
	// Every line comes from cmd itself, about a second apart. A second shell that
	// has to start is silent for as long as the runner is loaded, and a test about
	// the idle clock must not be a test about PowerShell's startup: the echo that
	// used to open this command re-armed the clock and then left two seconds of
	// PowerShell starting to look like a stall.
	return `ping -n 2 127.0.0.1 >nul & for /l %i in (1,1,3) do @(echo x & ping -n 2 127.0.0.1 >nul)`
}

func quietAfterFirstCommand() string { return "echo first & ping -n 3 127.0.0.1 >nul" }
