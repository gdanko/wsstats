package stats

import "github.com/shirou/gopsutil/v3/host"

type HostInformation struct {
	Information  *host.InfoStat         `json:"information"`
	Temperatures []host.TemperatureStat `json:"temperatures"`
	Users        []host.UserStat        `json:"users"`
}

func GetHostInformation() (hostInformation HostInformation, err error) {
	var (
		hostInfo  *host.InfoStat
		hostTemps []host.TemperatureStat
		hostUsers []host.UserStat
	)
	hostInfo, err = host.Info()
	if err != nil {
		return hostInformation, err
	}

	hostTemps, err = host.SensorsTemperatures()
	if err != nil {
		return hostInformation, err
	}

	hostUsers, err = host.Users()
	if err != nil {
		return hostInformation, err
	}

	hostInformation = HostInformation{
		Information:  hostInfo,
		Temperatures: hostTemps,
		Users:        hostUsers,
	}

	return hostInformation, nil
}
