package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func catalogo(t *testing.T) *Catalog {
	t.Helper()
	for _, r := range []string{
		"../../overtrace-library/modules/catalog.json",
		"overtrace-library/modules/catalog.json",
	} {
		if b, err := os.ReadFile(r); err == nil {
			var c Catalog
			if err := json.Unmarshal(b, &c); err != nil {
				t.Fatal(err)
			}
			return &c
		}
	}
	t.Skip("catalog.json no encontrado")
	return nil
}

// preset mínimo que pasa todo, sobre el que se introducen fallos de uno en uno
func base() *Preset {
	anio := 1979
	return &Preset{
		ID: "x", SchemaVersion: 1,
		Song:        Song{Title: "t", Artist: "a", ReleaseYear: &anio},
		Section:     "solo",
		Performance: Performance{Kind: "studio", Year: &anio},
		Stages: []Stage{{Role: "amp", Lanes: []Lane{{Nodes: []Node{{
			ModuleURI: "overtrace:nam", Confidence: "inferred",
			Note: ptr("deducido del tempo"),
		}}}}}},
		Provenance: Provenance{Confidence: "inferred", Sources: []Source{}},
	}
}
func ptr(s string) *string { return &s }

func tiene(in *Informe, regla string) bool {
	for _, h := range in.Hallazgos {
		if h.Regla == regla {
			return true
		}
	}
	return false
}
func accion(in *Informe, regla string, ci bool) Accion {
	for _, h := range in.Hallazgos {
		if h.Regla == regla {
			if ci {
				return h.EnCI
			}
			return h.EnAgente
		}
	}
	return Avisar
}

func TestElPresetBaseNoTieneHallazgos(t *testing.T) {
	in := Validar(base(), catalogo(t))
	if len(in.Hallazgos) != 0 {
		t.Fatalf("el preset base debería pasar limpio: %+v", in.Hallazgos)
	}
}

func TestModuloInexistenteSeRechaza(t *testing.T) {
	p := base()
	p.Stages[0].Lanes[0].Nodes[0].ModuleURI = "overtrace:inventado"
	if !tiene(Validar(p, catalogo(t)), "Módulos") {
		t.Fatal("no caza un módulo que no existe")
	}
}

func TestElCompensadorNoPuedeIrEnUnPreset(t *testing.T) {
	// §7.3: es etapa implícita del motor y ajuste local, no contenido de preset.
	p := base()
	p.Stages[0].Lanes[0].Nodes = append(p.Stages[0].Lanes[0].Nodes,
		Node{ModuleURI: "overtrace:pickup", Confidence: "estimated", Note: ptr("x")})
	if !tiene(Validar(p, catalogo(t)), "Compensador") {
		t.Fatal("deja colar overtrace:pickup en el preset")
	}
}

func TestParametroFueraDeRango(t *testing.T) {
	p := base()
	p.Stages[0].Lanes[0].Nodes[0].Params = map[string]float64{"input_level": 99}
	in := Validar(p, catalogo(t))
	if !tiene(in, "Parámetros") {
		t.Fatal("no caza un parámetro fuera de rango")
	}
	// en el agente rechaza; en CI degrada, para no tirar el trabajo de nadie
	if accion(in, "Parámetros", false) != Rechazar {
		t.Fatal("en el agente debería rechazar")
	}
	if accion(in, "Parámetros", true) != Degradar {
		t.Fatal("en CI debería degradar, no rechazar")
	}
}

func TestAnacronismoSoloSobreGearNoSobreAsset(t *testing.T) {
	// La lección de la v2.4: el asset es el fichero que aproxima, no el equipo.
	p := base()
	a2010 := 2010
	p.Stages[0].Lanes[0].Nodes[0].Gear = &Gear{Name: "Pedal de 2010", Year: &a2010}
	if !tiene(Validar(p, catalogo(t)), "Anacronismo") {
		t.Fatal("no caza equipo posterior a la interpretación")
	}

	// mismo caso pero en el asset: NO debe ser un hallazgo de anacronismo
	q := base()
	q.Stages[0].Lanes[0].Nodes[0].Asset = &Asset{
		Kind: "nam", DisplayName: "captura moderna",
		SourceURL: "https://ejemplo.com/x", Match: "similar"}
	if tiene(Validar(q, catalogo(t)), "Anacronismo") {
		t.Fatal("valida el año del asset: eso rechazaría presets correctos")
	}
}

func TestElRigDeGiraPosteriorAlDiscoEsValido(t *testing.T) {
	// El caso que motivó separar `performance` de `song` (§6.1).
	p := base()
	gira := 1994
	p.Performance = Performance{Kind: "live", Year: &gira}
	a1973 := 1973
	p.Stages[0].Lanes[0].Nodes[0].Gear = &Gear{Name: "Hiwatt DR103", Year: &a1973}
	if tiene(Validar(p, catalogo(t)), "Anacronismo") {
		t.Fatal("rechaza un rig de gira de 1994 con equipo de 1973: es correcto")
	}
}

func TestDocumentedSinFuenteSeDegrada(t *testing.T) {
	p := base()
	p.Stages[0].Lanes[0].Nodes[0].Confidence = "documented"
	in := Validar(p, catalogo(t))
	if !tiene(in, "Fuentes") {
		t.Fatal("no caza 'documented' sin cita propia")
	}
	if accion(in, "Fuentes", false) != Degradar {
		t.Fatal("debe DEGRADAR, no rechazar: el preset se conserva y el semáforo dice la verdad")
	}
}

func TestSinAmplificadorSeRechaza(t *testing.T) {
	p := base()
	p.Stages[0].Lanes[0].Nodes[0].ModuleURI = "overtrace:delay"
	if !tiene(Validar(p, catalogo(t)), "Topología") {
		t.Fatal("un rig sin amplificador debería rechazarse")
	}
}

func TestVariosAmplificadoresSonValidos(t *testing.T) {
	// ADR-07: un rig de varios amplis es esperable, no un error.
	p := base()
	l := p.Stages[0].Lanes[0]
	p.Stages[0].Lanes = []Lane{l, l, l}
	if tiene(Validar(p, catalogo(t)), "Topología") {
		t.Fatal("rechaza un wet/dry/wet: es justo lo que el producto reconstruye")
	}
}

func TestCitaDePortadaSeRechaza(t *testing.T) {
	// Consecuencia 3 de §9.5.
	p := base()
	n := &p.Stages[0].Lanes[0].Nodes[0]
	n.Confidence = "documented"
	n.Gear = &Gear{Name: "Hiwatt DR103"}
	n.Sources = []Source{{URL: "https://www.gilmourish.com/"}}
	in := Validar(p, catalogo(t))
	ValidarCitas(p, in, false)
	if !tiene(in, "Citas verificables") {
		t.Fatal("acepta la portada de un sitio como fuente")
	}
}

func TestPaginaConcretaSeAcepta(t *testing.T) {
	p := base()
	n := &p.Stages[0].Lanes[0].Nodes[0]
	n.Confidence = "documented"
	n.Gear = &Gear{Name: "Hiwatt DR103"}
	n.Sources = []Source{{URL: "https://www.gilmourish.com/gear/the-black-strat/"}}
	in := Validar(p, catalogo(t))
	ValidarCitas(p, in, false)
	if tiene(in, "Citas verificables") {
		t.Fatal("rechaza una URL que sí apunta a una página concreta")
	}
}

func TestQuoteDemasiadoLargo(t *testing.T) {
	p := base()
	n := &p.Stages[0].Lanes[0].Nodes[0]
	n.Confidence = "documented"
	n.Sources = []Source{{URL: "https://x.com/a", Quote: strings.Repeat("a", 201)}}
	if !tiene(Validar(p, catalogo(t)), "Citas") {
		t.Fatal("no aplica el límite de 200 caracteres de R-15")
	}
}
