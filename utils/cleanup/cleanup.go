package cleanup

import (
	"MYSQL-orchestration-API/config"
	"fmt"
	"os/exec"
)

func RemoveSlave(serverconfig *config.ServerConfig, SlaveID string) {
	container := fmt.Sprintf("%s-server", SlaveID)
	cmd := exec.Command("docker", "rm", "-f", container)
	cmd.Env = append(cmd.Env, "DOCKER_HOST=npipe:////./pipe/docker_engine")
	output, err := cmd.CombinedOutput()
	fmt.Printf("Cleaned %s", string(output))
	if err != nil {
		// return
	}
	volume := fmt.Sprintf("%s-data", SlaveID)
	cmd = exec.Command("docker", "volume", "rm", volume)
	cmd.Env = append(cmd.Env, "DOCKER_HOST=npipe:////./pipe/docker_engine")
	output, err = cmd.CombinedOutput()

	if err != nil {
		// return
	}
	fmt.Printf("Cleaned %s", string(output))
	fmt.Printf("Cleaned %s", serverconfig)
}

func Cleanup(Slaves map[string]*config.ServerConfig, Master map[string]*config.ServerConfig) {
	Cleaner := make(map[string]*config.ServerConfig)

	for key, serverconfig := range Slaves {
		Cleaner[key] = serverconfig
		delete(Slaves, key)
	}
	for key, serverconfig := range Master {
		Cleaner[key] = serverconfig
		delete(Master, key)
	}
	for key, serverconfig := range Cleaner {
		// value := Slaves[key]
		RemoveSlave(serverconfig, key)
	}
}
