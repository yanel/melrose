package main

import (
	"fmt"

	"github.com/emicklei/melrose/audio"
	"github.com/robertkrimen/otto"
)

var Otto = otto.New()

var Audio *audio.Device

func main() {
	setup()
	fmt.Println(help())
	Audio = new(audio.Device)
	Audio.Open()
	Audio.LoadSounds()
	defer Audio.Close()
	loop()
}

func help() string {
	return `
	melrose
	
	play("C#5 E_ 2F G A#")
	tempo(150)
	chord("C") -> "C E"
	scale("C") -> "C D E F G A B"
	pitch("C" , -1) -> "B3"
	reverse("C D E") -> "E D C"
	repeat("C",5)
	rotate("C D E",-1) -> "D E C"
`
}

func setup() {
	Otto.Set("play", playAllSequences)
	Otto.Set("tempo", tempo)
	Otto.Set("chord", chord)
	Otto.Set("scale", scale)
	Otto.Set("pitch", pitch)
	Otto.Set("reverse", reverse)
	Otto.Set("repeat", repeat)
	Otto.Set("rotate", rotate)
}