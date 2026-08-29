package main

// Validadores semánticos de §9.3.
//
// Un schema válido no garantiza contenido cierto: éste es el bucle que
// convierte R-02 de "esperemos que el modelo no alucine" en algo medible.
//
// Cada regla declara DOS acciones: la del agente (donde reintentar cuesta una
// llamada) y la de CI (donde rechazar tira el trabajo de una persona).

import (
	"fmt"
	"math"
	"strings"
)

type Accion int

const (
	Rechazar Accion = iota
	Degradar
	Avisar
)

func (a Accion) String() string {
	return [...]string{"RECHAZAR", "DEGRADAR", "AVISAR"}[a]
}

type Hallazgo struct {
	Regla    string
	Donde    string // ruta legible: "stages[1].lanes[0].nodes[2]"
	Mensaje  string
	EnAgente Accion
	EnCI     Accion
}

type Informe struct {
	Hallazgos []Hallazgo
}

func (in *Informe) add(regla, donde, msg string, agente, ci Accion) {
	in.Hallazgos = append(in.Hallazgos, Hallazgo{regla, donde, msg, agente, ci})
}

func (in *Informe) Cuenta(a Accion, ci bool) int {
	n := 0
	for _, h := range in.Hallazgos {
		act := h.EnAgente
		if ci {
			act = h.EnCI
		}
		if act == a {
			n++
		}
	}
	return n
}

// Validar aplica las reglas de §9.3. No muta el preset: informa.
func Validar(p *Preset, cat *Catalog) *Informe {
	in := &Informe{}
	mods := map[string]Module{}
	for _, m := range cat.Modules {
		mods[m.URI] = m
	}

	// ── año de referencia para el anacronismo ──
	anio := 0
	if p.Performance.Year != nil {
		anio = *p.Performance.Year
	} else if p.Song.ReleaseYear != nil {
		anio = *p.Song.ReleaseYear
	}

	nAmps, nNam, nCarriles := 0, 0, 0

	for si, st := range p.Stages {
		if len(st.Lanes) == 0 {
			in.add("Estructura", fmt.Sprintf("stages[%d]", si),
				"la etapa no tiene carriles", Rechazar, Rechazar)
		}

		// ── Nivel de una etapa en paralelo ──
		//
		// El sumador NO compensa por número de carriles: `gain_db` significa
		// el nivel del carril y nada más. Eso es deliberado —un wet/dry/wet
		// existe para que el seco quede intacto— pero deja el nivel total en
		// manos de quien escribe el preset. Aquí es donde se avisa.
		//
		// Se suman las amplitudes, no las potencias: carriles que salen de la
		// MISMA señal están correlacionados, así que en el peor caso se suman
		// linealmente. Es el caso que hay que vigilar.
		if len(st.Lanes) > 1 {
			suma := 0.0
			for _, ln := range st.Lanes {
				g := 0.0
				if ln.GainDB != nil {
					g = *ln.GainDB
				}
				suma += math.Pow(10, g/20)
			}
			if db := 20 * math.Log10(suma); db > 6 {
				in.add("Carriles", fmt.Sprintf("stages[%d]", si),
					fmt.Sprintf("los %d carriles pueden sumar hasta +%.1f dB: si salen de la "+
						"misma señal van correlacionados y se suman enteros. Baja gain_db "+
						"o cuenta con que el limitador trabaje", len(st.Lanes), db),
					Degradar, Degradar)
			}
		}

		for li, ln := range st.Lanes {
			nCarriles++
			donde := fmt.Sprintf("stages[%d].lanes[%d]", si, li)

			if len(ln.Nodes) == 0 {
				in.add("Estructura", donde, "el carril no tiene nodos", Rechazar, Rechazar)
			}
			if ln.Pan != nil && (*ln.Pan < -1 || *ln.Pan > 1) {
				in.add("Estructura", donde,
					fmt.Sprintf("pan %.2f fuera de [-1,1]", *ln.Pan), Rechazar, Degradar)
			}
			if ln.GainDB != nil && (*ln.GainDB < -60 || *ln.GainDB > 12) {
				in.add("Estructura", donde,
					fmt.Sprintf("gain_db %.1f fuera de [-60,12]", *ln.GainDB), Rechazar, Degradar)
			}
			// Carriles: etapa de un solo carril debe ser neutra
			if len(st.Lanes) == 1 {
				if (ln.Pan != nil && *ln.Pan != 0) || (ln.GainDB != nil && *ln.GainDB != 0) {
					in.add("Carriles", donde,
						"etapa de un solo carril con pan/gain no neutros: mezcla invisible para el usuario",
						Degradar, Degradar)
				}
			}

			for ni, nd := range ln.Nodes {
				dn := fmt.Sprintf("%s.nodes[%d]", donde, ni)
				m, ok := mods[nd.ModuleURI]

				// ── Preferencias con mandos propios (ADR-12, forma larga) ──
				//
				// Un nodo puede llevar los mandos del plugin concreto que
				// prefiere, además de los de la capacidad. Eso da control fino
				// donde ese plugin existe, y no viaja: en otra máquina se
				// ignoran. La trampa es apoyarse SOLO en ellos — entonces el
				// preset suena como quien lo hizo quería en su ordenador y en
				// ningún otro, sin que nadie se entere.
				for pi, pref := range nd.Prefer {
					dp := fmt.Sprintf("%s.prefer[%d]", dn, pi)
					if pref.URI == "" {
						in.add("Preferencias", dp,
							"una preferencia sin uri no apunta a nada", Rechazar, Rechazar)
						continue
					}
					if !strings.HasPrefix(pref.URI, "lv2:") &&
						!strings.HasPrefix(pref.URI, "clap:") {
						in.add("Preferencias", dp,
							fmt.Sprintf("%q no lleva prefijo de formato: se espera "+
								"'lv2:<uri>' o 'clap:<id>'", pref.URI), Rechazar, Degradar)
					}
					if len(pref.Params) > 0 && len(nd.Params) == 0 {
						in.add("Preferencias", dp,
							fmt.Sprintf("este nodo solo ajusta los %d mando(s) propios de "+
								"%s y ninguno de la capacidad: en cualquier máquina que no "+
								"tenga ese plugin sonará con los valores por defecto",
								len(pref.Params), pref.URI), Degradar, Degradar)
					}
				}

				// ── Módulos ──
				if !ok {
					in.add("Módulos", dn,
						fmt.Sprintf("module_uri %q no existe en el catálogo", nd.ModuleURI),
						Rechazar, Rechazar)
				} else {
					if m.IsAmp {
						nAmps++
					}
					if m.URI == "overtrace:nam" {
						nNam++
					}
					// ── Compensador: propiedad del motor, no del preset (§7.3) ──
					if m.EngineOwned {
						in.add("Compensador", dn,
							fmt.Sprintf("%s es etapa implícita del motor y ajuste local; no puede ir en un preset", m.URI),
							Rechazar, Rechazar)
					}
					// ── Parámetros en rango ──
					for k, v := range nd.Params {
						pr, existe := m.Params[k]
						if !existe {
							in.add("Parámetros", dn,
								fmt.Sprintf("%s no declara el parámetro %q", m.URI, k),
								Rechazar, Degradar)
							continue
						}
						if v < pr.Min || v > pr.Max {
							in.add("Parámetros", dn,
								fmt.Sprintf("%s=%g fuera de [%g,%g] %s", k, v, pr.Min, pr.Max, pr.Unit),
								Rechazar, Degradar)
						}
					}
				}

				// ── Portabilidad: todo extended necesita fallback al núcleo ──
				if ok && m.Tier != "core" {
					if nd.FallbackURI == nil {
						in.add("Portabilidad", dn,
							fmt.Sprintf("%s no es del núcleo y no declara fallback_uri", nd.ModuleURI),
							Rechazar, Rechazar)
					} else if fm, fok := mods[*nd.FallbackURI]; !fok || fm.Tier != "core" {
						in.add("Portabilidad", dn,
							fmt.Sprintf("fallback_uri %q no es un módulo del núcleo", *nd.FallbackURI),
							Rechazar, Rechazar)
					}
				}
				if !ok && nd.FallbackURI != nil {
					if fm, fok := mods[*nd.FallbackURI]; fok && fm.Tier == "core" {
						// módulo extendido desconocido pero con fallback válido: aceptable
						in.Hallazgos = in.Hallazgos[:len(in.Hallazgos)-1] // retira el "no existe"
						in.add("Módulos", dn,
							fmt.Sprintf("%s no está en el catálogo local, pero declara fallback al núcleo", nd.ModuleURI),
							Avisar, Avisar)
					}
				}

				// ── Anacronismo: SOLO sobre gear, nunca sobre asset ni module_uri ──
				if nd.Gear != nil && nd.Gear.Year != nil && anio > 0 {
					if *nd.Gear.Year > anio {
						in.add("Anacronismo", dn,
							fmt.Sprintf("%q es de %d, posterior a la interpretación (%d)",
								nd.Gear.Name, *nd.Gear.Year, anio),
							Rechazar, Degradar)
					}
				}

				// ── Fuentes: documented sin cita propia se degrada ──
				if nd.Confidence == "documented" && len(nd.Sources) == 0 {
					in.add("Fuentes", dn,
						"confidence 'documented' sin ninguna entrada en sources: se degrada a 'inferred'",
						Degradar, Degradar)
				}
				// ── Notas ──
				if (nd.Confidence == "inferred" || nd.Confidence == "estimated") &&
					len(nd.Sources) == 0 && (nd.Note == nil || *nd.Note == "") {
					in.add("Notas", dn,
						fmt.Sprintf("confidence %q sin fuentes y sin note: el usuario no sabe de dónde sale el valor",
							nd.Confidence),
						Avisar, Avisar)
				}
				// ── Assets: match distinto de exact NO es error ──
				if nd.Asset != nil && nd.Asset.Match != "" && nd.Asset.Match != "exact" {
					in.add("Assets", dn,
						fmt.Sprintf("la captura es %q respecto al gear declarado (informativo, no es un error)",
							nd.Asset.Match),
						Avisar, Avisar)
				}
				// ── Citas: límite de longitud (R-15) ──
				for _, s := range nd.Sources {
					if len(s.Quote) > 200 {
						in.add("Citas", dn,
							fmt.Sprintf("quote de %d caracteres, máximo 200", len(s.Quote)),
							Rechazar, Rechazar)
					}
				}
			}
		}
	}

	// ── Topología: al menos un amplificador ──
	if nAmps == 0 {
		in.add("Topología", "stages",
			"ningún módulo de amplificador en todo el rig", Rechazar, Rechazar)
	}
	// ── Coste ──
	if nNam > 3 {
		in.add("Coste", "stages",
			fmt.Sprintf("%d instancias NAM: preset exigente de CPU", nNam), Avisar, Avisar)
	}
	// ── Coherencia temporal ──
	if p.Performance.Year != nil && p.Song.ReleaseYear != nil {
		py, ry := *p.Performance.Year, *p.Song.ReleaseYear
		if p.Performance.Kind == "live" && py < ry {
			in.add("Coherencia temporal", "performance",
				fmt.Sprintf("directo de %d anterior a la publicación (%d)", py, ry), Avisar, Avisar)
		}
		if p.Performance.Kind == "studio" && py > ry {
			in.add("Coherencia temporal", "performance",
				fmt.Sprintf("grabación de estudio de %d posterior a la publicación (%d)", py, ry), Avisar, Avisar)
		}
	}
	// ── Fuentes del preset ──
	if p.Provenance.Confidence == "documented" && len(p.Provenance.Sources) == 0 {
		in.add("Fuentes", "provenance",
			"preset 'documented' sin fuentes", Degradar, Degradar)
	}
	return in
}
