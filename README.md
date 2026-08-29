# OverTrace · biblioteca

> *¿Qué cadena de sonido usó este artista en esta parte de esta canción?*

Este repositorio **no tiene código**: tiene la respuesta a esa pregunta, escrita
como datos verificables. Un preset describe un rig —pedales, amplificador,
pantalla, ruteo— y, para cada afirmación, de dónde sale.

El programa que lo toca vive aparte. Aquí solo está el catálogo, y se consume
por HTTPS anónimo: sin git, sin cuenta, sin token.

---

## Qué hay dentro

```
schema/preset.v1.schema.json   El contrato. Lo que un preset puede decir
presets/                       Un fichero por preset. Inmutables
modules/catalog.json           Las capacidades del núcleo y con qué se realizan
profiles/pickups.json          Curvas de compensación por tipo de pastilla
profiles/instruments.json      Guitarras conocidas y las posiciones de su selector
index.json                     Índice ligero: qué hay y cuánto cuesta, sin abrirlo todo
.github/workflows/             La CI, que es la única puerta de calidad
```

## Lo que hace distinto a este catálogo

**Cada afirmación lleva su procedencia.** Un nodo no dice solo «Big Muff»: dice
si eso está *documentado* con una fuente que se puede abrir, *deducido* del
contexto, o *estimado* a ojo. La interfaz lo pinta con un semáforo, y el
validador rechaza un `documented` sin fuentes.

Eso importa porque el fallo típico de reconstruir un rig con un modelo de
lenguaje no es inventarse un pedal imposible —eso se caza solo— sino afirmar
algo **plausible y falso** con total aplomo. Separar lo que se sabe de lo que se
supone no lo elimina, pero lo pone donde se ve.

**El preset pide capacidades, no plugins.** Un nodo dice «quiero saturación» y,
como mucho, con qué preferiría hacerlo. El motor resuelve a lo mejor que tenga
esa máquina y **dice qué consiguió**. Por eso el mismo fichero suena en Linux y
en Windows, con distinto acabado y sin huecos.

**Los presets son inmutables.** Adaptar uno a tu guitarra no lo edita: crea otro
con `parent_id`. Lo que el artista usó es un hecho y no cambia porque tú toques
otra cosa.

## Añadir un preset

Un PR, un preset, y la CI decide. No hay revisión humana: hay siete
comprobaciones mecánicas más los validadores semánticos, y el mismo binario que
usa el agente de investigación —no una reimplementación, porque si divergieran
el agente daría por bueno lo que el PR va a rechazar.

Ver [CONTRIBUTING.md](CONTRIBUTING.md).

## Estado

**Cuatro presets, y dos son la misma canción.** Es el riesgo que decide si esto
sirve para algo: un catálogo con procedencia y sin contenido no vale nada. Hasta
que no haya treinta o cuarenta canciones revisadas **a oído**, esto es una
demostración de que el formato funciona, no una biblioteca.

## Licencia

El contenido va bajo **CC-BY-NC 4.0** (ver [LICENSE](LICENSE)). El texto del CLA
para contribuciones está pendiente de redacción y revisión; hasta entonces, un
PR se acepta entendiendo que su contenido queda bajo la misma licencia.

El software que lee esto —motor, validador, herramientas— está en otro
repositorio y bajo otra licencia.
