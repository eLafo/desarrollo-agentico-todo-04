// Pide confirmación antes de enviar formularios marcados con data-confirm.
document.addEventListener("submit", function (event) {
  var form = event.target;
  if (form.dataset.confirm && !window.confirm(form.dataset.confirm)) {
    event.preventDefault();
  }

  if (!form.matches("[data-task-toggle]")) {
    return;
  }

  event.preventDefault();

  var task = form.closest(".task");
  var completed = task.classList.toggle("task-completed");
  var button = form.querySelector("button");
  var status = task.querySelector(".task-status");

  var label = completed ? "Reabrir" : "Completar";
  button.querySelector(".toggle-label").textContent = label;
  button.setAttribute("aria-pressed", String(completed));
  button.setAttribute("aria-label", label + " tarea: " + task.querySelector(".task-title").textContent);
  button.setAttribute("title", label + " tarea");

  if (completed && !status) {
    status = document.createElement("span");
    status.className = "task-status";
    status.textContent = "Completada";
    task.insertBefore(status, task.querySelector(".task-actions"));
  } else if (!completed && status) {
    status.remove();
  }

  fetch(form.action);
});

// Reordenación de la lista: arrastre con ratón o táctil desde el asa y botones
// Subir/Bajar para teclado. Cada movimiento se guarda en el servidor, de uno en uno.
(function () {
  var list = document.querySelector("[data-task-list]");
  if (!list) {
    return;
  }

  var status = document.querySelector("[data-reorder-status]");
  var savedOrder = taskIds();
  var queue = Promise.resolve();
  var generation = 0;
  var drag = null;

  list.classList.add("task-list-reorderable");

  function tasks() {
    return Array.prototype.filter.call(list.children, function (element) {
      return element.matches(".task");
    });
  }

  function taskIds() {
    return tasks().map(function (task) {
      return task.dataset.taskId;
    });
  }

  function announce(message, isError) {
    status.textContent = message;
    status.classList.toggle("sr-only", !isError);
    status.classList.toggle("field-error", Boolean(isError));
  }

  function announcePosition(task) {
    var all = tasks();
    var title = task.querySelector(".task-title").textContent;
    announce("Tarea «" + title + "» movida a la posición " + (all.indexOf(task) + 1) + " de " + all.length + ".");
  }

  // Vuelve a colocar las tareas en el orden indicado y conserva el foco.
  function restore(ids) {
    var active = document.activeElement;
    var byId = {};
    tasks().forEach(function (task) {
      byId[task.dataset.taskId] = task;
    });
    ids.forEach(function (id) {
      if (byId[id]) {
        list.appendChild(byId[id]);
      }
    });
    if (active && list.contains(active)) {
      active.focus();
    }
  }

  // Guarda la posición actual de la tarea con el mismo formulario que usa
  // la versión sin JavaScript. Las peticiones salen en orden y, si una falla,
  // se descartan las pendientes y la lista vuelve al último orden guardado.
  function save(task) {
    var form = task.querySelector("[data-task-move]");
    var body = new URLSearchParams(new FormData(form));
    body.set("position", String(tasks().indexOf(task) + 1));
    var order = taskIds();
    var current = generation;

    queue = queue.then(function () {
      if (current !== generation) {
        return;
      }
      // El servidor responde con una redirección 303; "manual" evita pedir
      // la página de destino y la convierte en una respuesta "opaqueredirect".
      return fetch(form.action, {
        method: "POST",
        body: body,
        credentials: "same-origin",
        redirect: "manual"
      })
        .then(function (response) {
          if (response.type !== "opaqueredirect" && !response.ok) {
            throw new Error("HTTP " + response.status);
          }
          savedOrder = order;
        })
        .catch(function () {
          generation += 1;
          cancelDrag();
          restore(savedOrder);
          announce("No se ha podido guardar el nuevo orden. Recarga la página e inténtalo de nuevo.", true);
        });
    });
  }

  // Botones Subir/Bajar: se mueve la tarea vecina y no la propia tarea, para
  // que el botón pulsado conserve el foco y se pueda repetir el movimiento.
  list.addEventListener("submit", function (event) {
    var form = event.target;
    if (!form.matches("[data-task-move]")) {
      return;
    }
    event.preventDefault();

    var task = form.closest(".task");
    var up = form.dataset.taskMove === "up";
    var neighbour = up ? task.previousElementSibling : task.nextElementSibling;
    if (!neighbour) {
      announce(up ? "La tarea ya es la primera de la lista." : "La tarea ya es la última de la lista.");
      return;
    }

    list.insertBefore(neighbour, up ? task.nextElementSibling : task);
    announcePosition(task);
    save(task);
  });

  // Arrastre con ratón, lápiz o dedo (Pointer Events) desde el asa. La captura
  // del puntero se hace sobre la lista, que nunca se mueve, para no perderla
  // al recolocar la tarea arrastrada.
  list.addEventListener("pointerdown", function (event) {
    var handle = event.target.closest("[data-drag-handle]");
    if (!handle || drag || !event.isPrimary || event.button !== 0) {
      return;
    }
    event.preventDefault();

    var task = handle.closest(".task");
    list.setPointerCapture(event.pointerId);
    drag = { task: task, pointerId: event.pointerId, startOrder: taskIds() };
    task.classList.add("task-dragging");
    document.body.classList.add("is-reordering");
  });

  list.addEventListener("pointermove", function (event) {
    if (!drag || event.pointerId !== drag.pointerId) {
      return;
    }
    event.preventDefault();

    var before = null;
    var others = tasks().filter(function (task) {
      return task !== drag.task;
    });
    for (var i = 0; i < others.length; i++) {
      var rect = others[i].getBoundingClientRect();
      if (event.clientY < rect.top + rect.height / 2) {
        before = others[i];
        break;
      }
    }
    if (drag.task.nextElementSibling !== before) {
      list.insertBefore(drag.task, before);
    }

    var edge = 48;
    if (event.clientY < edge) {
      window.scrollBy(0, -12);
    } else if (event.clientY > window.innerHeight - edge) {
      window.scrollBy(0, 12);
    }
  });

  function endDrag(commit) {
    var current = drag;
    drag = null;
    current.task.classList.remove("task-dragging");
    document.body.classList.remove("is-reordering");
    if (list.hasPointerCapture(current.pointerId)) {
      list.releasePointerCapture(current.pointerId);
    }

    var moved = taskIds().join() !== current.startOrder.join();
    if (!commit) {
      restore(current.startOrder);
    } else if (moved) {
      announcePosition(current.task);
      save(current.task);
    }
  }

  function cancelDrag() {
    if (drag) {
      endDrag(false);
    }
  }

  list.addEventListener("pointerup", function (event) {
    if (drag && event.pointerId === drag.pointerId) {
      endDrag(true);
    }
  });

  list.addEventListener("pointercancel", function (event) {
    if (drag && event.pointerId === drag.pointerId) {
      endDrag(false);
    }
  });

  document.addEventListener("keydown", function (event) {
    if (event.key === "Escape") {
      cancelDrag();
    }
  });
})();
