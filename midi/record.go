package midi

import (
	"fmt"
	"github.com/emicklei/melrose/core"
	"time"

	"github.com/rakyll/portmidi"
)

// Record is part of melrose.AudioDevice
func (m *Midi) Record(deviceID int, stopAfterInactivity time.Duration) (*core.Recording, error) {
	rec := core.NewRecording()
	in, err := portmidi.NewInputStream(portmidi.DeviceID(deviceID), 1024) // buffer
	if err != nil {
		return rec, err
	}
	defer in.Close()

	midiDeviceInfo := portmidi.Info(portmidi.DeviceID(deviceID))
	info(fmt.Sprintf("recording from %s/%s ... [until %v silence]\n", midiDeviceInfo.Interface, midiDeviceInfo.Name, stopAfterInactivity))

	ch := in.Listen()
	timeout := time.NewTimer(stopAfterInactivity)
	needsReset := false
	now := time.Now()
	for {
		if needsReset {
			timeout.Reset(stopAfterInactivity)
			needsReset = false
		}
		select {
		case each := <-ch: // depending on the device, this may not block and other events are received
			when := now.Add(time.Duration(each.Timestamp) * time.Millisecond)
			if each.Status == noteOn {
				print(core.MIDItoNote(int(each.Data1), 1.0))
				rec.Add(core.NewNoteChange(true, each.Data1, each.Data2), when)
				needsReset = true
				continue
			}
			if each.Status != noteOff {
				continue
			}
			// note off
			needsReset = true
			rec.Add(core.NewNoteChange(false, each.Data1, each.Data2), when)
			if !timeout.Stop() {
				<-timeout.C
			}
		case <-timeout.C:
			goto done
		}
	}
done:
	info(fmt.Sprintf("\nstopped after %v of inactivity\n", stopAfterInactivity))
	return rec, nil
}

// TODO compute duration
func (m *Midi) eventToNote(start, end portmidi.Event) core.Note {
	return core.MIDItoNote(int(start.Data1), int(start.Data2))
}
