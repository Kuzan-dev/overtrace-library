package main

// overtrace-validate — valida presets contra el schema y los validadores
// semánticos de §9.3.
//
//   uso:  overtrace-validate [--ci] <preset.json> [preset.json ...]
//
//   --ci   aplica las acciones de CI (degradar en vez de rechazar donde
//          corresponde). Sin la bandera, aplica las del agente.

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const (
	rojo  = "\033[31m"
	ambar = "\033[33m"
	verde = "\033[32m"
	gris  = "\033[90m"
	neg   = "\033[1m"
	fin   = "\033[0m"
)

func main() {
	ci := flag.Bool("ci", false, "aplicar las acciones de CI en lugar de las del agente")
	catPath := flag.String("catalogo", "", "ruta a modules/catalog.json (autodetecta si se omite)")
	citas := flag.Bool("verificar-citas", false, "descargar cada URL de sources y comprobar que respalda el gear (necesita red)")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "uso: overtrace-validate [--ci] <preset.json> ...")
		os.Exit(2)
	}

	cat, err := cargarCatalogo(*catPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "no se pudo cargar el catálogo: %v\n", err)
		os.Exit(2)
	}

	modo := "agente"
	if *ci {
		modo = "CI"
	}
	fmt.Printf("\n%sovertrace-validate%s · modo %s · catálogo con %d módulos\n",
		neg, fin, modo, len(cat.Modules))

	totalRech, totalDeg, totalAvi, ficheros := 0, 0, 0, 0
	for _, ruta := range flag.Args() {
		ficheros++
		datos, err := os.ReadFile(ruta)
		if err != nil {
			fmt.Printf("\n%s✘ %s%s — no se puede leer: %v\n", rojo, ruta, fin, err)
			totalRech++
			continue
		}
		var p Preset
		if err := json.Unmarshal(datos, &p); err != nil {
			fmt.Printf("\n%s✘ %s%s — JSON inválido: %v\n", rojo, ruta, fin, err)
			totalRech++
			continue
		}

		in := Validar(&p, cat)
		ValidarCitas(&p, in, *citas)
		r := in.Cuenta(Rechazar, *ci)
		d := in.Cuenta(Degradar, *ci)
		a := in.Cuenta(Avisar, *ci)
		totalRech += r
		totalDeg += d
		totalAvi += a

		icono, color := "✔", verde
		if r > 0 {
			icono, color = "✘", rojo
		} else if d > 0 {
			icono, color = "~", ambar
		}
		fmt.Printf("\n%s%s %s%s  %s — %s (%s)\n", color, icono, filepath.Base(ruta), fin,
			p.Song.Artist, p.Song.Title, p.Section)

		for _, h := range in.Hallazgos {
			act := h.EnAgente
			if *ci {
				act = h.EnCI
			}
			c := gris
			switch act {
			case Rechazar:
				c = rojo
			case Degradar:
				c = ambar
			}
			fmt.Printf("   %s%-9s%s %s%-14s%s %s\n   %s%s%s\n",
				c, act, fin, neg, h.Regla, fin, h.Mensaje, gris, h.Donde, fin)
		}
		if len(in.Hallazgos) == 0 {
			fmt.Printf("   %ssin hallazgos%s\n", gris, fin)
		}
	}

	fmt.Printf("\n%s──────────────────────────────────────────%s\n", gris, fin)
	fmt.Printf("  %d fichero(s) · %s%d rechazar%s · %s%d degradar%s · %d avisar\n\n",
		ficheros, rojo, totalRech, fin, ambar, totalDeg, fin, totalAvi)
	if totalRech > 0 {
		os.Exit(1)
	}
}

func cargarCatalogo(ruta string) (*Catalog, error) {
	if ruta == "" {
		for _, c := range []string{
			"overtrace-library/modules/catalog.json",
			"../overtrace-library/modules/catalog.json",
			"../../overtrace-library/modules/catalog.json",
		} {
			if _, err := os.Stat(c); err == nil {
				ruta = c
				break
			}
		}
	}
	if ruta == "" {
		return nil, fmt.Errorf("no encontrado (usa --catalogo)")
	}
	datos, err := os.ReadFile(ruta)
	if err != nil {
		return nil, err
	}
	var c Catalog
	return &c, json.Unmarshal(datos, &c)
}
