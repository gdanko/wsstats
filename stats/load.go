package stats

import "github.com/shirou/gopsutil/v3/load"

func GetLoadAverages() (loadAverages *load.AvgStat, err error) {
	loadAverages, err = load.Avg()
	if err != nil {
		return loadAverages, err
	}
	return loadAverages, nil
}
