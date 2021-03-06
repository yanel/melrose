package main

import (
	"flag"
	"fmt"
	"github.com/emicklei/melrose/core"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/emicklei/melrose/dsl"
	"github.com/emicklei/melrose/notify"
	"github.com/emicklei/melrose/server"
	"github.com/peterh/liner"
)

var (
	version   = "dev"
	verbose   = flag.Bool("v", false, "verbose logging")
	inputFile = flag.String("i", "", "read expressions from a file")
	httpPort  = flag.String("http", ":8118", "address on which to listen for HTTP requests")

	history = ".melrose.history"
)

func main() {
	welcome()
	flag.Parse()

	// set audio
	currentDevice := setupAudio("midi")
	defer currentDevice.Close()

	ctx := new(core.PlayContext)
	ctx.AudioDevice = currentDevice
	ctx.VariableStorage = dsl.NewVariableStore()
	ctx.LoopControl = core.NewBeatmaster(ctx, 120)

	// process file if given
	if len(*inputFile) > 0 {
		if err := processInputFile(ctx, *inputFile); err != nil {
			notify.Print(notify.Error(err))
			os.Exit(0)
		}
	}

	if len(*httpPort) > 0 {
		// start DSL server
		go server.NewLanguageServer(ctx, *httpPort).Start()
	}

	// start REPL
	line := liner.NewLiner()
	defer line.Close()
	defer tearDown(line, ctx)
	// TODO liner catches control+c
	//setupCloseHandler(line)
	setup(line)
	repl(line, ctx)
}

func welcome() {
	fmt.Println("\033[1;34mmelrōse\033[0m" + " - program your melodies")
}

func tearDown(line *liner.State, ctx core.Context) {
	//dsl.StopAllLoops(store)
	//control.Stop()

	ctx.Control().Reset()
	ctx.Device().Reset()
	if f, err := os.Create(history); err != nil {
		notify.Print(notify.Errorf("error writing history file:%v", err))
	} else {
		line.WriteHistory(f)
		f.Close()
	}
	fmt.Println("\033[1;34mmelrose\033[0m" + " sings bye!")
}

func setup(line *liner.State) {
	line.SetCtrlCAborts(true)
	line.SetWordCompleter(completeMe)
	if f, err := os.Open(history); err == nil {
		line.ReadHistory(f)
		f.Close()
	}
}

func repl(line *liner.State, ctx core.Context) {
	eval := dsl.NewEvaluator(ctx)
	ctx.Control().Start()
	for {
		entry, err := line.Prompt("𝄞 ")
		if err != nil {
			notify.Print(notify.Error(err))
			continue
		}
		entry = strings.TrimSpace(entry)
		if strings.HasPrefix(entry, ":") {
			// special case
			if entry == ":q" || entry == ":Q" {
				goto exit
			}
			args := strings.Split(entry, " ")
			if cmd, ok := lookupCommand(args[0]); ok {
				if msg := cmd.Func(ctx, args[1:]); msg != nil {
					notify.Print(msg)
				}
				continue
			}
		}
		if result, err := eval.EvaluateStatement(entry); err != nil {
			notify.Print(notify.Error(err))
			// even on error, add entry to history so we can edit/fix it
		} else {
			if result != nil {
				core.PrintValue(ctx, result)
			}
		}
		line.AppendHistory(entry)
	}
exit:
}

func processInputFile(ctx core.Context, inputFile string) error {
	data, err := ioutil.ReadFile(inputFile)
	if err != nil {
		notify.Print(notify.Errorf("unable to read file:%v", err))
		return nil
	}
	eval := dsl.NewEvaluator(ctx)
	_, err = eval.EvaluateProgram(string(data))
	return err
}

// setupCloseHandler creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS. We then handle this by calling
// our clean up procedure and exiting the program.
func setupCloseHandler(line *liner.State, ctx core.Context) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		tearDown(line, ctx)
		os.Exit(0)
	}()
}
