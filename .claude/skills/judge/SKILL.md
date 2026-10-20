---
name: judge
description: Juzga un material concreto (fragmentos del diff, textos de la interfaz, la traza de un cambio a un criterio) con una rúbrica fijada de antemano y deja un acta de juicio con cita literal, razonamiento por criterio, confianza y procedencia; o intenta refutar un hallazgo de otro juicio. Úsala para lo que un test no puede decidir. No la uses para lo que se puede automatizar.
allowed-tools: Read, Glob, Grep, Write, Bash(git diff *), Bash(git show *), Bash(git log *)
---

# Juzga con rúbrica

Un test trae su propia prueba: el código es la referencia y el resultado se repite. Un juicio no: la confianza tiene que venir del proceso, y el acta del juicio lo muestra. Hay dos modos: **juzgar** y **refutar**.

## Juzgar

### Entradas

- La afirmación que se juzga y su rúbrica, por su identificador en `docs/procedures/conventions.md` (`E1`, `F1`, `D1`…).
- El material, ya extraído: los fragmentos del diff con fichero y línea, las cadenas nuevas, la traza de un fragmento a un criterio. Si te pasan el diff entero, extrae tú lo que la rúbrica dice que juzga.
- El fichero donde escribir el acta del juicio, dentro de `.adw/`.

No recibes la justificación del implementador ni el razonamiento de otros jueces. Si la encuentras, no la uses.

### Procedimiento

1. Lee la rúbrica entera: sus niveles y sus ejemplos. Es la única referencia: no añadas criterios. Si ves algo importante que la rúbrica no cubre, anótalo aparte como observación para el humano, sin severidad.
2. Para cada criterio de la rúbrica que aplique al material:
   - **Enuncia** el criterio aplicado a este material, en una frase que se entiende sin la rúbrica: «La acción solo coordina», «El comentario de `TitleMaxLength` sigue siendo cierto».
   - **Cita** literalmente lo que juzgas, con fichero y línea. Comprueba que la cita existe tal cual (`git diff`, `Read`).
   - **Razona** por qué cumple o no cumple, también cuando cumple. «Lo veo bien» no es un razonamiento. Resume después el razonamiento en una frase, sin identificadores de afirmaciones ni de rúbricas.
   - **Decide** el nivel según la rúbrica: cumple, `nit`, `should-fix` o `blocker`. Un nivel que la rúbrica no prevé para ese criterio no existe.
   - **Da tu confianza:** alta si la cita basta para demostrarlo a cualquiera que lea la rúbrica; baja si depende de una interpretación o de contexto que no tienes.
3. No contradigas un resultado automatizado: un test en rojo no se declara aceptable, y un gate en verde no se discute con un juicio.
4. Escribe el acta del juicio y devuélvela.

### Acta del juicio

```markdown
### <afirmación> · <rúbrica> · <criterio>
- **Enunciado:** <el criterio aplicado a este material, en una frase>
- **Material:** `ruta:línea` — «cita literal»
- **Razonamiento:** …
- **En una frase:** <el razonamiento resumido>
- **Veredicto:** cumple | nit | should-fix | blocker
- **Confianza:** alta | baja — por qué
- **Procedencia:** rúbrica <id> de `docs/procedures/conventions.md` · modelo <modelo> · material <rango del diff o ficheros>
```

Un bloque por criterio y por fragmento juzgado. Al final, la lista de hallazgos (`nit`, `should-fix`, `blocker`) y las observaciones fuera de rúbrica.

El acta del juicio se queda en `.adw/` y el humano no la ve. Lo que se publica de ella es, por cada criterio, el enunciado, la cita y el razonamiento: en una frase si cumple y entero si no. Por eso esos tres campos se escriben para el humano: con palabras, sin los identificadores de la afirmación ni de la rúbrica.

## Refutar

Cada hallazgo `blocker` pasa por un juez distinto, sin contexto previo, que intenta tumbarlo antes de que cuente.

### Entradas

- El hallazgo: su rúbrica, su criterio, su cita, su razonamiento y su severidad.
- La rúbrica y el material citado. No recibes nada más del primer juez.

### Procedimiento

1. Comprueba que la cita existe tal cual en el cambio y que el fichero y la línea son los que dice.
2. Comprueba que el criterio aplica a ese material y que la severidad es la que la rúbrica da a ese caso.
3. Busca lo que lo justificaría: un criterio de la especificación que lo exija, un patrón del repositorio que lo haga igual, un contexto que el primer juez no vio.
4. Decide:
   - **Refutado:** la cita no existe, el criterio no aplica, la rúbrica le da otra severidad o hay una justificación que lo desmonta. Di cuál.
   - **Sostenido:** nada de lo anterior. Di qué intentaste.
5. Da tu confianza en el hallazgo, alta o baja, con el mismo criterio que al juzgar.

### Resultado

```markdown
### Refutación de <afirmación> · <rúbrica> · <criterio>
- **Comprobado:** cita · criterio · severidad · justificaciones buscadas
- **Resultado:** sostenido | refutado — por qué
- **Confianza:** alta | baja
- **Procedencia:** modelo <modelo>
```

Un `blocker` juzgado solo rechaza un cambio si sale sostenido y con confianza alta. Si no, va al humano como punto a revisar.

## Límites

- Juzga solo el material y con la rúbrica que te dan.
- No cambies el código ni publiques nada: el acta del juicio es tu único resultado.
