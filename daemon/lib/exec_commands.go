package lib

import (
	"io/ioutil"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func ExecCommand(input ...string) error {

	cmd := exec.Command(input[0], input[0:]...)
	stdout, errOUT := cmd.StdoutPipe()
	if errOUT != nil {
		LogE(errOUT).Warning("Impossible to obtain the STDOUT pipe")
	}
	stderr, errERR := cmd.StderrPipe()
	if errERR != nil {
		LogE(errERR).Warning("Impossible to obtain the STDERR pipe")
	}
	err := cmd.Start()
	if err != nil {
		LogE(err).Error("Error in starting the command")
	}

	slurpOut, errOUT := ioutil.ReadAll(stdout)
	if errOUT != nil {
		LogE(errOUT).Warning("Impossible to read the STDOUT")
	}
	slurpErr, errERR := ioutil.ReadAll(stderr)
	if errERR != nil {
		LogE(errERR).Warning("Impossible to read the STDERR")
	}

	err = cmd.Wait()
	if err != nil {
		LogE(err).Error("Error in executing the command")
		Log().WithFields(log.Fields{"pipe": "STDOUT"}).Info(slurpOut)
		Log().WithFields(log.Fields{"pipe": "STDERR"}).Info(slurpErr)
		return err
	}
	return nil
}
