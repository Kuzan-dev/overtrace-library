# Contribuir al catálogo de OverTrace

**No hace falta que uses git ni GitHub.** La aplicación lo hace por ti: pulsas
**Publicar** y ella abre el *pull request*, con tu cuenta, por debajo (ADR-10).
Esta página es para quien quiera entender qué pasa o contribuir a mano.

## Qué se puede añadir

Un preset por PR, como fichero nuevo en `presets/<uuid>.json`.

**El repositorio es append-only.** Nadie edita el preset de otro: mejorar uno
ajeno crea uno nuevo con `parent_id` apuntando al original (ADR-05). Así los
conflictos de merge son estructuralmente imposibles.

## Qué comprueba la CI

Todo automático. Si pasa, **el PR se mergea solo**; no espera a nadie.

| Comprobación | Rechaza si |
| --- | --- |
| Alcance | Toca algo fuera de `presets/` |
| Append-only | Modifica o borra un fichero existente |
| Cantidad | Añade más de un preset |
| Nombre | El fichero no se llama `<id>.json` con el `id` de dentro |
| Schema | No valida contra `schema/preset.v1.schema.json` |
| Semántica | Falla un validador de §9.3 con acción *rechazar* |
| Portabilidad | Usa un módulo fuera del núcleo sin `fallback_uri` |
| CLA | Tu usuario no consta como firmante |

Si algo falla, el bot te dice **qué** y **por qué** en lenguaje llano. No hay
que adivinar.

## Sobre la honestidad de los datos

Es lo que hace que este catálogo valga algo, así que va en serio:

* **`confidence` por nodo, no por preset.** Que el amplificador esté documentado
  no significa que la posición de las perillas lo esté. Casi nunca lo está.
* **`documented` exige una cita propia** que apunte a **una página concreta**.
  La portada de un sitio no es una fuente: `ejemplo.com/` no vale,
  `ejemplo.com/entrevista-1994/` sí.
* **Si no lo sabes, `inferred` o `estimated` con un `note`** que explique de
  dónde sale el número. Un dato honesto marcado como inferido vale más que uno
  falso marcado como documentado — y esto no es retórica: se midió que un
  modelo genera datos plausibles y falsos que ningún validador puede cazar
  (§9.5). El semáforo es la única defensa.

## El CLA

Una pantalla, una sola vez, dentro de la propia aplicación.

**Conservas tu copyright.** Lo que concedes es una licencia para usar,
modificar, sublicenciar y explotar comercialmente tu aportación, incluido el
entrenamiento de modelos. Sin eso, un catálogo de miles de presets de cientos
de personas sería inexplotable: habría que pedir permiso preset a preset, para
siempre.

Detalle completo en §14.3 del documento de arquitectura.
