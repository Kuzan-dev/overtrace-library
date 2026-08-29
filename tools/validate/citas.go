package main

// Validador «Citas verificables» — §9.5, consecuencia 2 y 3 de la prueba de R-02.
//
// La prueba demostró que un nodo puede afirmar `documented`, llevar una URL
// real, y aun así ser falso: la URL apuntaba a la raíz de un sitio que trata
// del tema en general, no a una página que respalde ESA afirmación.
//
// Dos comprobaciones, con costes muy distintos:
//   · URL de raíz  → sin red, instantánea, siempre activa
//   · Contenido    → con red, lenta, tras --verificar-citas

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var etiquetasHTML = regexp.MustCompile(`(?s)<(script|style)[^>]*>.*?</(script|style)>|<[^>]+>`)

// esRaiz decide si una URL es la portada de un sitio en lugar de una página
// concreta. Una cita tiene que apuntar a algo que se pueda leer y comprobar.
func esRaiz(cruda string) bool {
	u, err := url.Parse(cruda)
	if err != nil {
		return false
	}
	ruta := strings.Trim(u.Path, "/")
	if ruta == "" && u.RawQuery == "" {
		return true
	}
	// dominios de portada disfrazada: /index.html, /es/, /home
	switch strings.ToLower(ruta) {
	case "index.html", "index.php", "home", "es", "en":
		return u.RawQuery == ""
	}
	return false
}

// normaliza deja un texto comparable: minúsculas, sin acentos comunes,
// sin puntuación, espacios colapsados.
func normaliza(s string) string {
	r := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ñ", "n",
		"Á", "a", "É", "e", "Í", "i", "Ó", "o", "Ú", "u", "Ñ", "n",
		"'", " ", "’", " ", "\"", " ", "-", " ", "_", " ", ".", " ", ",", " ",
	)
	s = r.Replace(strings.ToLower(s))
	return strings.Join(strings.Fields(s), " ")
}

// menciona comprueba si el cuerpo respalda el nombre del equipo. No exige la
// cadena entera —"Hiwatt Custom 100 DR103" rara vez aparece literal— sino que
// aparezcan sus palabras significativas.
func menciona(cuerpo, nombre string) bool {
	c := normaliza(cuerpo)
	palabras := strings.Fields(normaliza(nombre))
	sig := 0
	aciertos := 0
	for _, p := range palabras {
		if len(p) < 3 {
			continue // "de", "the", números de dos cifras
		}
		sig++
		if strings.Contains(c, p) {
			aciertos++
		}
	}
	if sig == 0 {
		return true
	}
	// dos tercios de las palabras significativas presentes
	return float64(aciertos)/float64(sig) >= 0.66
}

type verificador struct {
	cliente *http.Client
	cache   map[string]string
}

func nuevoVerificador() *verificador {
	return &verificador{
		cliente: &http.Client{Timeout: 12 * time.Second},
		cache:   map[string]string{},
	}
}

func (v *verificador) cuerpo(u string) (string, error) {
	if c, ok := v.cache[u]; ok {
		return c, nil
	}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "overtrace-validate/0.1 (verificación de citas)")
	resp, err := v.cliente.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	datos, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", err
	}
	texto := etiquetasHTML.ReplaceAllString(string(datos), " ")
	v.cache[u] = texto
	return texto, nil
}

// ValidarCitas recorre los nodos `documented` y comprueba sus fuentes.
// Si verificarContenido es false, solo aplica la regla de URL de raíz.
func ValidarCitas(p *Preset, in *Informe, verificarContenido bool) {
	var v *verificador
	if verificarContenido {
		v = nuevoVerificador()
	}

	revisar := func(donde string, conf string, srcs []Source, gear *Gear) {
		if conf != "documented" {
			return
		}
		for _, s := range srcs {
			if esRaiz(s.URL) {
				in.add("Citas verificables", donde,
					fmt.Sprintf("%q es la portada del sitio, no una página que respalde la afirmación", s.URL),
					Rechazar, Avisar)
				continue
			}
			if !verificarContenido || gear == nil || gear.Name == "" {
				continue
			}
			cuerpo, err := v.cuerpo(s.URL)
			if err != nil {
				in.add("Citas verificables", donde,
					fmt.Sprintf("%q no se pudo leer: %v", s.URL, err), Avisar, Avisar)
				continue
			}
			if !menciona(cuerpo, gear.Name) {
				in.add("Citas verificables", donde,
					fmt.Sprintf("%q responde, pero su texto no menciona %q", s.URL, gear.Name),
					Rechazar, Avisar)
			}
		}
	}

	if p.SourceInstrument != nil {
		g := &Gear{Name: ""}
		if p.SourceInstrument.Guitar != nil {
			g.Name = *p.SourceInstrument.Guitar
		}
		revisar("source_instrument", p.SourceInstrument.Confidence, p.SourceInstrument.Sources, g)
	}
	for si, st := range p.Stages {
		for li, ln := range st.Lanes {
			for ni, nd := range ln.Nodes {
				revisar(fmt.Sprintf("stages[%d].lanes[%d].nodes[%d]", si, li, ni),
					nd.Confidence, nd.Sources, nd.Gear)
			}
		}
	}
}
