//go:build unix

package shellcmd_test

func quietCommand() string { return "sleep 2" }
func chattyCommand() string {
	return "echo a; sleep 0.15; echo b; sleep 0.15; echo c; sleep 0.15; echo d"
}
func quietAfterFirstCommand() string { return "echo first; sleep 2" }
