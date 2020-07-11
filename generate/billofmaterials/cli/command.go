package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bom-maker/bomcsv"
	"github.com/bom-maker/mouser"
	"github.com/bom-maker/output"
	cli "github.com/jawher/mow.cli"
)

var (
	// MouserAPIKey is the key used to make API calls
	MouserAPIKey *string
)

// Process processes command line args
func Process(appName, appDesc, appVersion string) {
	app := cli.App(appName, appDesc)
	app.Spec = "-k"

	app.Version("v version", fmt.Sprintf("%s version %s", appName, appVersion))

	MouserAPIKey = app.StringOpt("k mouser-api-key", "", "Mouser API key to use")

	app.Command("generate", "Generate a BOM file from CSV (stdin)", generate)
	app.Command("cart", "Create a Mouser cart from CSV (stdin)", cart)

	app.Action = func() {
		fmt.Println("Please choose a command arg!")
	}

	app.Run(os.Args)
}

func generate(cmd *cli.Cmd) {
	cmd.Spec = "-o [-s] [-t] [-l]"

	outputMode := cmd.StringOpt("o output", "csv", "Output mode [csv,html]")
	inputSeparator := cmd.StringOpt("s in-sep", ";", "CSV separator used to read from stdin")
	outputSeparator := cmd.StringOpt("t out-sep", ";", "Separator used to write to stdout")
	htmlTitle := cmd.StringOpt("l html-title", "BOM - @time", "HTML title to use for HTML generation, use @time to include date and time")

	cmd.Action = func() {
		// Params checking
		if *outputMode != "csv" && *outputMode != "html" {
			fmt.Fprintf(os.Stderr, "Unknown output %s\n", *outputMode)
			return
		}
		if len(*inputSeparator) != 1 || len(*outputSeparator) != 1 {
			fmt.Fprintln(os.Stderr, "Separators must be exactly one character long")
			return
		}

		csvParts, err := bomcsv.ReadCSVPartsFrom(os.Stdin, *inputSeparator)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%+v", err)
			return
		}

		parts := createUberParts(*MouserAPIKey, csvParts)

		switch *outputMode {
		case "csv":
			out := output.CSV{
				Parts:     parts,
				Separator: *outputSeparator,
			}
			err = out.Write(os.Stdout)
		case "html":
			now := time.Now()
			title := *htmlTitle
			title = strings.Replace(title, "@time", now.Format("2006-01-02T15:04:05-0700"), 1)
			out := output.HTML{
				Parts: parts,
				Title: title,
			}
			err = out.Write(os.Stdout)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "An error occurred, err=%v", err)
		}
	}
}

func cart(cmd *cli.Cmd) {
	cmd.Spec = "-c [-s] [-m]"

	inputSeparator := cmd.StringOpt("s in-sep", ";", "CSV separator used to read from stdin")
	mouserCartAPIKey := cmd.StringOpt("c cart-api-key", "", "Mouser cart/order API key to use")
	multiplier := cmd.IntOpt("m mult", 1, "Multiply each item added to the cart by this multiplier")

	cmd.Action = func() {
		if len(*inputSeparator) != 1 {
			fmt.Fprintln(os.Stderr, "Separators must be exactly one character long")
			return
		}

		// Getting csv parts
		csvParts, err := bomcsv.ReadCSVPartsFrom(os.Stdin, *inputSeparator)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%+v", err)
			return
		}

		// Getting uber parts
		parts := createUberParts(*MouserAPIKey, csvParts)

		// Creating a cart
		api := mouser.NewAPI(*mouserCartAPIKey)
		results, err := api.InsertItemsInCart("", parts, *multiplier)
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error occurred creating cart, err=%v\n", err)
			return
		}

		// Checking results
		resultsJSON, err := json.Marshal(results)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot marshal results as json, err=%v", err)
			return
		}
		fmt.Printf("%v", string(resultsJSON))
	}
}
