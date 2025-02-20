package ui

import (
	"fmt"

	"github.com/fatih/color"
)

type Printer struct {
	cyan  *color.Color
	white *color.Color
	gray  *color.Color
	green *color.Color
}

func NewPrinter() *Printer {
	return &Printer{
		cyan:  color.New(color.FgCyan),
		white: color.New(color.FgWhite),
		gray:  color.New(color.FgHiBlack),
		green: color.New(color.FgGreen),
	}
}

func (p *Printer) PrintService(index int, name string, ports string) {
	p.cyan.Printf("[%d] ", index)
	p.white.Printf("%s ", name)
	p.gray.Printf("→ ")
	p.green.Printf("%s\n", ports)
}

func (p *Printer) PrintPort(index int, portInfo string) {
	p.cyan.Printf("[%d] ", index)
	p.green.Printf("%s\n", portInfo)
}

func (p *Printer) PrintForward(localPort, remotePort int) {
	fmt.Printf("\n")
	p.white.Printf("🚀 Forwarding port ")
	p.cyan.Printf("%d ", localPort)
	p.gray.Printf("→ ")
	p.cyan.Printf("%d\n", remotePort)
}

func (p *Printer) PrintLocalURL(port int) {
	p.white.Printf("💻 Local URL: ")
	p.green.Printf("http://localhost:%d\n\n", port)
}

func (p *Printer) PrintExit() {
	p.gray.Println("⌨️  Press Ctrl+C to exit")
}

func PrintError(format string, a ...interface{}) {
	color.Red("❌ "+format, a...)
}
