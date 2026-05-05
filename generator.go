package main

import (
	"slices"
	"strings"
)

type GenerateRequest struct {
	Preset      string   `json:"preset"`      // "indicator" | "strategy" | "library" | "custom"
	Declaration string   `json:"declaration"` // "indicator" | "strategy" | "library"
	Sections    []string `json:"sections"`    // ordered list of section keys (only used when preset = "custom")
	Mode        string   `json:"mode"`        // "prefilled" | "minimal"
}

var presetSections = map[string][]string{
	"indicator": {"version", "declaration", "constants", "inputs", "functions", "logic", "plots"},
	"strategy":  {"version", "declaration", "constants", "inputs", "functions", "logic", "strategy_calls"},
	"library":   {"version", "declaration", "types", "functions"},
}

func (r *GenerateRequest) resolve() {
	if r.Declaration == "" {
		r.Declaration = "indicator"
	}
	if r.Mode == "" {
		r.Mode = "prefilled"
	}
	if r.Preset != "" && r.Preset != "custom" {
		if sections, ok := presetSections[r.Preset]; ok {
			r.Sections = sections
			r.Declaration = r.Preset
		}
	}
}

var sectionCode = map[string]string{
	"version": `// This Pine Script® code is subject to the terms of the Mozilla Public License 2.0 at https://mozilla.org/MPL/2.0/
// © NexoMechanics

//@version=6`,

	"declaration_indicator": `indicator("My Indicator", overlay = false)`,
	"declaration_strategy":  `strategy("My Strategy", overlay = false, initial_capital = 10000, default_qty_type = strategy.percent_of_equity, default_qty_value = 100)`,
	"declaration_library":   `// @description A library of reusable Pine Script functions and types.
library("MyLibrary", overlay = false)`,

	"imports": `import TradingView/ta/7 as ta2`,

	"constants": `BULL_COLOR = color.green
BEAR_COLOR = color.red`,

	"inputs": `i_length = input.int(14, "Length", minval = 1)
i_source = input.source(close, "Source")`,

	"inputs_strategy": `i_length = input.int(14, "Length", minval = 1)
i_source = input.source(close, "Source")`,

	"types": `type Bar
    float o = open
    float h = high
    float l = low
    float c = close`,

	"types_library": `export type Bar
    float o = open
    float h = high
    float l = low
    float c = close`,

	"functions": `calcMA(src = close, len = 14) =>
    ta.sma(src, len)`,

	"functions_library": `// @function Calculates a simple moving average over a given source and length.
// @param src The source series to average. Defaults to close.
// @param len The number of bars to average over. Defaults to 14.
// @returns The simple moving average of src over len bars.
export calcMA(float src = close, int len = 14) =>
    ta.sma(src, len)`,

	"logic": `maValue    = calcMA(i_source, i_length)
bullSignal = ta.crossover(close, maValue)
bearSignal = ta.crossunder(close, maValue)`,

	"strategy_calls": `if bullSignal
    strategy.entry("Long", strategy.long)
if bearSignal
    strategy.entry("Short", strategy.short)`,

	"plots": `plot(maValue, "MA", color = close > maValue ? BULL_COLOR : BEAR_COLOR, linewidth = 2)
plotshape(bullSignal, "Bull", shape.triangleup, location.belowbar, BULL_COLOR, size = size.small)
plotshape(bearSignal, "Bear", shape.triangledown, location.abovebar, BEAR_COLOR, size = size.small)`,

	"alerts": `alertcondition(bullSignal, "Bull Signal", "Bull signal detected")
alertcondition(bearSignal, "Bear Signal", "Bear signal detected")`,

	"alerts_strategy": `if bullSignal
    alert("Bull signal detected", alert.freq_once_per_bar_close)

if bearSignal
    alert("Bear signal detected", alert.freq_once_per_bar_close)`,
}

var sectionCodeMinimal = map[string]string{
	"imports":        "",
	"constants":      "",
	"inputs":         "",
	"types":          "",
	"functions":      "",
	"logic":          "",
	"strategy_calls": "",
	"plots":          "",
	"alerts":         "",
}

var sectionHeaders = map[string]string{
	"version":        "",
	"declaration":    "",
	"imports":        "// Imports {",
	"constants":      "// Constants {",
	"inputs":         "// Inputs {",
	"types":          "// Types {",
	"functions":      "// Functions & Methods {",
	"logic":          "// Logic {",
	"strategy_calls": "// Strategy {",
	"plots":          "// Plots {",
	"alerts":         "// Alerts {",
}

var sectionFooters = map[string]string{
	"imports":        "// }",
	"constants":      "// }",
	"inputs":         "// }",
	"types":          "// }",
	"functions":      "// }",
	"logic":          "// }",
	"strategy_calls": "// }",
	"plots":          "// }",
	"alerts":         "// }",
}

func hasSection(sections []string, key string) bool {
	return slices.Contains(sections, key)
}

func generate(req GenerateRequest) string {
	minimal := req.Mode == "minimal"

	sections := req.Sections
	useConstantColors := hasSection(sections, "constants")
	useInputs := hasSection(sections, "inputs")
	useFunctions := hasSection(sections, "functions")
	useLogic := hasSection(sections, "logic")

	var parts []string

	for _, section := range sections {
		var code string

		if section == "version" {
			code = sectionCode["version"]
		} else if section == "declaration" {
			key := "declaration_" + req.Declaration
			if c, ok := sectionCode[key]; ok {
				if minimal {
					switch req.Declaration {
					case "strategy":
						code = `strategy("My Strategy")`
					case "library":
						code = `library("MyLibrary")`
					default:
						code = `indicator("My Indicator")`
					}
				} else {
					code = c
				}
			} else {
				code = sectionCode["declaration_indicator"]
			}
		} else if minimal {
			if _, ok := sectionCodeMinimal[section]; !ok {
				continue
			}
		} else {
			var ok bool
			code, ok = sectionCode[section+"_"+req.Declaration]
			if !ok {
				code, ok = sectionCode[section]
			}
			if !ok {
				continue
			}
		}

		if section != "constants" && !useConstantColors {
			code = strings.ReplaceAll(code, "BULL_COLOR", "color.green")
			code = strings.ReplaceAll(code, "BEAR_COLOR", "color.red")
		}
		if section == "functions" && useInputs {
			code = strings.ReplaceAll(code, "src = close, len = 14", "src, len")
		}
		if section == "functions_library" && useInputs {
			code = strings.ReplaceAll(code, "float src = close, int len = 14", "float src, int len")
		}
		if section == "logic" {
			if useFunctions && useInputs {
				// calcMA(i_source, i_length) — already in template
			} else if useFunctions && !useInputs {
				code = strings.ReplaceAll(code, "calcMA(i_source, i_length)", "calcMA()")
			} else if !useFunctions && useInputs {
				code = strings.ReplaceAll(code, "calcMA(i_source, i_length)", "ta.sma(i_source, i_length)")
			} else {
				code = strings.ReplaceAll(code, "calcMA(i_source, i_length)", "ta.sma(close, 14)")
			}
		} else if section == "plots" || section == "alerts" || section == "strategy_calls" {
			if !useInputs {
				code = strings.ReplaceAll(code, "i_source", "close")
				code = strings.ReplaceAll(code, "i_length", "14")
			}
			if !useLogic {
				src := "close"
				len := "14"
				if useInputs {
					src = "i_source"
					len = "i_length"
				}
				maExpr := "ta.sma(" + src + ", " + len + ")"
				code = strings.ReplaceAll(code, "maValue", maExpr)
				code = strings.ReplaceAll(code, "bullSignal", "ta.crossover(close, "+maExpr+")")
				code = strings.ReplaceAll(code, "bearSignal", "ta.crossunder(close, "+maExpr+")")
			}
		} else if section != "inputs" && section != "inputs_strategy" && !useInputs {
			code = strings.ReplaceAll(code, "i_source", "close")
			code = strings.ReplaceAll(code, "i_length", "14")
		}

		var block []string

		if header, ok := sectionHeaders[section]; ok && header != "" {
			block = append(block, header)
		}

		if code != "" {
			block = append(block, code)
		}

		if footer, ok := sectionFooters[section]; ok {
			if len(block) > 0 {
				block = append(block, "")
			}
			block = append(block, footer)
		}

		if len(block) > 0 {
			parts = append(parts, strings.Join(block, "\n"))
		}
	}

	return strings.Join(parts, "\n\n")
}
