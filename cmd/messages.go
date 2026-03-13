package cmd

import (
	"errors"
	"fmt"
)

var errNoOptions = errors.New("no options available")

func printCommandCanceled() {
	printSystem("Command canceled")
}

func printInfo(msg string) {
	printTagged("INFO", msg, colorBlue)
}

func printWarning(msg string) {
	printTagged("WARNING", msg, colorYellow)
}

func printError(msg string) {
	printTagged("ERROR", msg, colorRed)
}

func printSystem(msg string) {
	printTagged("SYSTEM", msg, colorRed)
}

func printTagged(tag, msg, color string) {
	fmt.Println(colorize(fmt.Sprintf("\n  [%s] %s", tag, msg), color))
}

func printWarningBox(msg string) {
	text := "\n  [WARNING] " + msg + "  \n"
	fmt.Println(colorize(text, colorYellow))
}

func printInfoBox(msg string) {
	text := "\n  [INFO] " + msg + "  \n"
	fmt.Println(colorize(text, colorBlue))
}

func printInfoCompact(msg string) {
	printTaggedCompact("INFO", msg, colorBlue)
}

func printTaggedCompact(tag, msg, color string) {
	fmt.Println(colorize(fmt.Sprintf("[%s] %s", tag, msg), color))
}
