package main

import (
	"github.com/sirupsen/logrus"

	manager "github.com/DataDog/ebpf-manager"
)

var m = &manager.Manager{
	Probes: []*manager.Probe{
		{
			ProbeIdentificationPair: manager.ProbeIdentificationPair{
				UID:          "MyFirstHook",
				EBPFSection:  "kprobe/vfs_mkdir",
				EBPFFuncName: "kprobe_vfs_mkdir",
			},
		},
		{
			ProbeIdentificationPair: manager.ProbeIdentificationPair{
				UID:          "",
				EBPFSection:  "kretprobe/mkdir",
				EBPFFuncName: "kretprobe_mkdir",
			},
			SyscallFuncName: "mkdir",
			KProbeMaxActive: 100,
		},
	},
	PerfMaps: []*manager.PerfMap{
		{
			Map: manager.Map{
				Name: "my_constants",
			},
			PerfMapOptions: manager.PerfMapOptions{
				DataHandler: myDataHandler,
			},
		},
	},
}

// myDataHandler - Perf event data handler
func myDataHandler(cpu int, data []byte, perfmap *manager.PerfMap, manager *manager.Manager) {
	myConstant := ByteOrder.Uint64(data[0:8])
	logrus.Printf("received: CPU:%d my_constant:%d", cpu, myConstant)
}

var editors = []manager.ConstantEditor{
	{
		Name:          "my_constant",
		Value:         uint64(100),
		FailOnMissing: true,
		ProbeIdentificationPairs: []manager.ProbeIdentificationPair{
			{UID: "MyFirstHook", EBPFSection: "kprobe/vfs_mkdir", EBPFFuncName: "kprobe_vfs_mkdir"},
		},
	},
	{
		Name:          "my_constant",
		Value:         uint64(555),
		FailOnMissing: true,
		ProbeIdentificationPairs: []manager.ProbeIdentificationPair{
			{UID: "", EBPFSection: "kprobe/vfs_mkdir", EBPFFuncName: "kprobe_vfs_mkdir"},
		},
	},
	{
		Name:                     "unused_constant",
		Value:                    uint64(555),
		ProbeIdentificationPairs: []manager.ProbeIdentificationPair{},
	},
}

func main() {
	// Prepare manager options
	options := manager.Options{ConstantEditors: editors}

	// Initialize the manager
	if err := m.InitWithOptions(recoverAssets(), options); err != nil {
		logrus.Fatal(err)
	}

	// Start the manager
	if err := m.Start(); err != nil {
		logrus.Fatal(err)
	}
	logrus.Println("eBPF programs running, head over to /sys/kernel/debug/tracing/trace_pipe to see them in action.")

	// Demo
	logrus.Println("INITIAL PROGRAMS")
	if err := trigger(); err != nil {
		_ = m.Stop(manager.CleanAll)
		logrus.Fatal(err)
	}
	if err := demoClone(); err != nil {
		_ = m.Stop(manager.CleanAll)
		logrus.Fatal(err)
	}
	if err := demoAddHook(); err != nil {
		_ = m.Stop(manager.CleanAll)
		logrus.Fatal(err)
	}

	// Close the manager
	if err := m.Stop(manager.CleanAll); err != nil {
		logrus.Fatal(err)
	}
}
