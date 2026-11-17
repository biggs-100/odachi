//go:build unix

package shellcmd_test

func quietCommand() string { return "sleep 2" }
func chattyCommand() string {
	return "echo a; for i in 1 2 3 4 5 6 7 8 9 10; do sleep 0.3; echo x; done"
}
func quietAfterFirstCommand() string { return "echo first; sleep 2" }
