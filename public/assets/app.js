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

// Reordenación manual: arrastre con ratón o táctil desde el asa y botones Subir/Bajar.
(function () {
  var list = document.querySelector("[data-task-list]");
  if (!list) {
    return;
  }

  var feedback = document.querySelector("[data-reorder-feedback]");
  var savedOrder = taskIds();
  var saving = false;
  var pending = false;
  var drag = null;

  list.classList.add("task-list-reorderable");

  function tasks() {
    return Array.prototype.filter.call(list.children, function (element) {
      return element.classList.contains("task");
    });
  }

  function taskIds() {
    return tasks().map(function (task) {
      return task.dataset.taskId;
    });
  }

  function sameOrder(a, b) {
    return a.length === b.length && a.every(function (id, index) {
      return id === b[index];
    });
  }

  function announce(message, isError) {
    feedback.textContent = message;
    feedback.classList.toggle("sr-only", !isError);
    feedback.classList.toggle("field-error", Boolean(isError));
  }

  function announcePosition(task) {
    var all = tasks();
    var title = task.querySelector(".task-title").textContent;
    announce("Tarea «" + title + "» movida a la posición " + (all.indexOf(task) + 1) + " de " + all.length + ".");
  }

  // Coloca task justo antes de target (null = al final) moviendo solo sus vecinas,
  // para que la tarea movida no salga del DOM y conserve el foco y la captura del puntero.
  function placeBefore(task, target) {
    if (target === task || task.nextElementSibling === target) {
      return;
    }
    var movingDown = target === null || Boolean(task.compareDocumentPosition(target) & Node.DOCUMENT_POSITION_FOLLOWING);
    while (task.nextElementSibling !== target) {
      if (movingDown) {
        list.insertBefore(task.nextElementSibling, task);
      } else {
        list.insertBefore(task.previousElementSibling, task.nextElementSibling);
      }
    }
  }

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

  function save() {
    if (saving) {
      pending = true;
      return;
    }
    var ids = taskIds();
    if (sameOrder(ids, savedOrder)) {
      return;
    }
    saving = true;
    fetch(list.dataset.reorderUrl, {
      method: "PUT",
      credentials: "same-origin",
      headers: {
        "Accept": "application/json",
        "Content-Type": "application/json",
        "X-CSRF-Token": list.dataset.csrfToken
      },
      body: JSON.stringify({ task_ids: ids })
    })
      .then(function (response) {
        if (!response.ok) {
          throw new Error("HTTP " + response.status);
        }
        savedOrder = ids;
      })
      .catch(function () {
        pending = false;
        restore(savedOrder);
        announce("No se pudo guardar el nuevo orden. Recarga la página e inténtalo de nuevo.", true);
      })
      .finally(function () {
        saving = false;
        if (pending) {
          pending = false;
          save();
        }
      });
  }

  // Alternativa con teclado (y lectores de pantalla): botones Subir/Bajar.
  list.addEventListener("click", function (event) {
    var button = event.target.closest("[data-move]");
    if (!button) {
      return;
    }
    var task = button.closest(".task");
    var up = button.dataset.move === "up";
    var neighbour = up ? task.previousElementSibling : task.nextElementSibling;
    if (!neighbour) {
      announce(up ? "La tarea ya es la primera de la lista." : "La tarea ya es la última de la lista.");
      return;
    }
    placeBefore(task, up ? neighbour : neighbour.nextElementSibling);
    announcePosition(task);
    save();
  });

  // Arrastre con ratón, lápiz o táctil (Pointer Events) desde el asa.
  list.addEventListener("pointerdown", function (event) {
    var handle = event.target.closest("[data-drag-handle]");
    if (!handle || drag || !event.isPrimary || event.button !== 0) {
      return;
    }
    event.preventDefault();
    var task = handle.closest(".task");
    handle.setPointerCapture(event.pointerId);
    drag = { task: task, pointerId: event.pointerId, startOrder: taskIds() };
    task.classList.add("task-dragging");
    document.body.classList.add("is-reordering");
  });

  list.addEventListener("pointermove", function (event) {
    if (!drag || event.pointerId !== drag.pointerId) {
      return;
    }
    event.preventDefault();
    var target = null;
    var others = tasks().filter(function (task) {
      return task !== drag.task;
    });
    for (var i = 0; i < others.length; i++) {
      var rect = others[i].getBoundingClientRect();
      if (event.clientY < rect.top + rect.height / 2) {
        target = others[i];
        break;
      }
    }
    placeBefore(drag.task, target);

    var edge = 48;
    if (event.clientY < edge) {
      window.scrollBy(0, -12);
    } else if (event.clientY > window.innerHeight - edge) {
      window.scrollBy(0, 12);
    }
  });

  function endDrag(event, commit) {
    if (!drag || event.pointerId !== drag.pointerId) {
      return;
    }
    var current = drag;
    drag = null;
    current.task.classList.remove("task-dragging");
    document.body.classList.remove("is-reordering");
    if (!commit) {
      restore(current.startOrder);
      return;
    }
    if (!sameOrder(taskIds(), current.startOrder)) {
      announcePosition(current.task);
      save();
    }
  }

  list.addEventListener("pointerup", function (event) {
    endDrag(event, true);
  });

  list.addEventListener("pointercancel", function (event) {
    endDrag(event, false);
  });
})();
