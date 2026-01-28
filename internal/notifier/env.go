package notifier

import "os"

func LookupEnv(k string) (string, bool) {
	return os.LookupEnv(k)
}

