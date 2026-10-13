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

// Reordena las tareas arrastrando el asa (ratón, táctil o lápiz) o con el teclado.
(function () {
  var list = document.querySelector("[data-task-list]");
  var orderForm = document.querySelector("form[data-task-order]");
  if (!list || !orderForm) {
    return;
  }

  var status = document.querySelector("[data-task-order-status]");
  var errorBox = document.querySelector("[data-task-order-error]");
  var EDGE = 48;
  var SCROLL_STEP = 16;
  var drag = null;
  var saving = false;
  var saveQueued = false;

  function items() {
    return Array.prototype.slice.call(list.children);
  }

  function currentIds() {
    return items().map(function (item) {
      return item.dataset.taskId;
    });
  }

  var confirmedIds = currentIds();

  function sameOrder(a, b) {
    return a.join(",") === b.join(",");
  }

  function middle(element) {
    var rect = element.getBoundingClientRect();
    return rect.top + rect.height / 2;
  }

  function announce(message) {
    status.textContent = message;
  }

  function announcePosition(item) {
    var all = items();
    announce("Tarea «" + item.querySelector(".task-title").textContent + "» movida a la posición " +
      (all.indexOf(item) + 1) + " de " + all.length + ".");
  }

  // Mueve la tarea un puesto desplazando a su vecina: la tarea no sale del DOM,
  // así conserva el foco del teclado y la captura del puntero.
  function shift(item, direction) {
    if (direction < 0 && item.previousElementSibling) {
      list.insertBefore(item.previousElementSibling, item.nextElementSibling);
      return true;
    }
    if (direction > 0 && item.nextElementSibling) {
      list.insertBefore(item.nextElementSibling, item);
      return true;
    }
    return false;
  }

  // Lleva la tarea al principio (direction < 0) o al final (direction > 0).
  function shiftToEdge(item, direction) {
    var moved = false;
    while (shift(item, direction)) {
      moved = true;
    }
    return moved;
  }

  // Cierra un movimiento ya hecho en el DOM: actualiza los botones de mover,
  // mantiene la tarea a la vista, lo anuncia y guarda el orden.
  function commitMove(item) {
    syncMoveButtons();
    item.scrollIntoView({ block: "nearest" });
    announcePosition(item);
    save();
  }

  function setUnavailable(button, unavailable) {
    if (unavailable) {
      button.setAttribute("aria-disabled", "true");
    } else {
      button.removeAttribute("aria-disabled");
    }
  }

  // «Mover al principio» no está disponible en la primera tarea, ni «Mover al final» en la última.
  function syncMoveButtons() {
    var all = items();
    all.forEach(function (item, index) {
      setUnavailable(item.querySelector('[data-task-move="first"]'), index === 0);
      setUnavailable(item.querySelector('[data-task-move="last"]'), index === all.length - 1);
    });
  }

  function applyOrder(ids) {
    var focused = document.activeElement;
    ids.forEach(function (id) {
      var item = list.querySelector('[data-task-id="' + id + '"]');
      if (item) {
        list.appendChild(item);
      }
    });
    syncMoveButtons();
    if (focused && list.contains(focused)) {
      focused.focus();
    }
  }

  function showError() {
    errorBox.textContent = "No se ha podido guardar el nuevo orden. Recarga la página e inténtalo de nuevo.";
    errorBox.hidden = false;
  }

  function hideError() {
    errorBox.hidden = true;
    errorBox.textContent = "";
  }

  // Guarda el orden completo. Solo hay una petición en curso; los cambios que llegan
  // mientras tanto se envían después con el orden más reciente.
  function save() {
    if (saving) {
      saveQueued = true;
      return;
    }
    saving = true;
    var ids = currentIds();
    var body = new URLSearchParams(new FormData(orderForm));
    ids.forEach(function (id) {
      body.append("task_ids", id);
    });

    fetch(orderForm.action, {
      method: "POST",
      body: body,
      credentials: "same-origin",
      headers: { Accept: "application/json" }
    })
      .then(function (response) {
        if (!response.ok) {
          throw new Error("HTTP " + response.status);
        }
        confirmedIds = ids;
        hideError();
      })
      .catch(function () {
        saveQueued = false;
        if (drag) {
          stopDrag();
        }
        applyOrder(confirmedIds);
        showError();
      })
      .then(function () {
        saving = false;
        if (saveQueued) {
          saveQueued = false;
          save();
        }
      });
  }

  function stopDrag() {
    var current = drag;
    drag = null;
    current.item.classList.remove("task-dragging");
    list.classList.remove("task-list-dragging");
    if (current.handle.hasPointerCapture(current.pointerId)) {
      current.handle.releasePointerCapture(current.pointerId);
    }
    return current;
  }

  function cancelDrag() {
    if (!drag) {
      return;
    }
    var current = stopDrag();
    if (!sameOrder(current.startIds, currentIds())) {
      applyOrder(current.startIds);
      announce("Movimiento cancelado.");
    }
  }

  list.addEventListener("pointerdown", function (event) {
    var handle = event.target.closest("[data-task-drag-handle]");
    if (!handle || drag || !event.isPrimary || event.button !== 0) {
      return;
    }
    event.preventDefault();
    var item = handle.closest("[data-task-id]");
    drag = { item: item, handle: handle, pointerId: event.pointerId, startIds: currentIds() };
    handle.setPointerCapture(event.pointerId);
    item.classList.add("task-dragging");
    list.classList.add("task-list-dragging");
  });

  list.addEventListener("pointermove", function (event) {
    if (!drag || event.pointerId !== drag.pointerId) {
      return;
    }
    var y = event.clientY;
    var moved = true;
    while (moved) {
      var previous = drag.item.previousElementSibling;
      var next = drag.item.nextElementSibling;
      if (previous && y < middle(previous)) {
        moved = shift(drag.item, -1);
      } else if (next && y > middle(next)) {
        moved = shift(drag.item, 1);
      } else {
        moved = false;
      }
    }
    if (y < EDGE) {
      window.scrollBy(0, -SCROLL_STEP);
    } else if (y > window.innerHeight - EDGE) {
      window.scrollBy(0, SCROLL_STEP);
    }
  });

  list.addEventListener("pointerup", function (event) {
    if (!drag || event.pointerId !== drag.pointerId) {
      return;
    }
    var current = stopDrag();
    if (!sameOrder(current.startIds, currentIds())) {
      commitMove(current.item);
    }
  });

  // lostpointercapture también llega tras pointerup, pero entonces drag ya es null.
  list.addEventListener("pointercancel", cancelDrag);
  list.addEventListener("lostpointercapture", cancelDrag);

  document.addEventListener("keydown", function (event) {
    if (drag && event.key === "Escape") {
      cancelDrag();
    }
  });

  list.addEventListener("keydown", function (event) {
    var handle = event.target.closest("[data-task-drag-handle]");
    if (!handle || drag || event.altKey || event.ctrlKey || event.metaKey || event.shiftKey) {
      return;
    }
    var item = handle.closest("[data-task-id]");
    var moved = false;
    if (event.key === "ArrowUp") {
      moved = shift(item, -1);
    } else if (event.key === "ArrowDown") {
      moved = shift(item, 1);
    } else if (event.key === "Home") {
      moved = shiftToEdge(item, -1);
    } else if (event.key === "End") {
      moved = shiftToEdge(item, 1);
    } else {
      return;
    }
    event.preventDefault();
    if (moved) {
      commitMove(item);
    }
  });

  // Botones «Mover al principio» y «Mover al final». El foco se queda en el botón,
  // porque la tarea no sale del DOM.
  list.addEventListener("click", function (event) {
    var button = event.target.closest("[data-task-move]");
    if (!button || drag || button.getAttribute("aria-disabled") === "true") {
      return;
    }
    var item = button.closest("[data-task-id]");
    if (shiftToEdge(item, button.dataset.taskMove === "first" ? -1 : 1)) {
      commitMove(item);
    }
  });
})();
