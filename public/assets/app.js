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

// Reordenación de tareas: arrastrar y soltar (HTML5) y Alt + flechas en el tirador.
(function () {
  var dragged = null;
  var idsAtStart = [];
  var saving = false;

  function listOf(el) {
    return el && el.closest ? el.closest("[data-task-list]") : null;
  }

  function currentIds(list) {
    return Array.prototype.map.call(list.querySelectorAll(":scope > .task"), function (li) {
      return li.dataset.taskId;
    });
  }

  function restore(list, ids) {
    ids.forEach(function (id) {
      var li = list.querySelector('.task[data-task-id="' + id + '"]');
      if (li) {
        list.appendChild(li);
      }
    });
  }

  function sameIds(a, b) {
    return a.length === b.length && a.every(function (id, i) { return id === b[i]; });
  }

  function saveOrder(list, previousIds) {
    var error = document.querySelector("[data-reorder-error]");
    var body = new URLSearchParams();
    body.append("_method", "PUT");
    body.append("authenticity_token", list.dataset.csrf);
    currentIds(list).forEach(function (id) { body.append("task_ids", id); });

    function fail() {
      restore(list, previousIds);
      if (error) {
        error.textContent = "No se pudo guardar el nuevo orden. Se ha restaurado la lista.";
        error.hidden = false;
      }
    }

    saving = true;
    return fetch(list.dataset.reorderUrl, {
      method: "POST",
      body: body,
      headers: { "Accept": "application/json" }
    }).then(function (response) {
      if (!response.ok) {
        fail();
      } else if (error) {
        error.hidden = true;
      }
    }, fail).then(function () {
      saving = false;
    });
  }

  document.addEventListener("dragstart", function (event) {
    var task = event.target.closest && event.target.closest("[data-task-list] .task");
    if (!task || saving) {
      if (task) { event.preventDefault(); }
      return;
    }
    dragged = task;
    idsAtStart = currentIds(listOf(task));
    task.classList.add("task-dragging");
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("text/plain", task.dataset.taskId);
  });

  document.addEventListener("dragover", function (event) {
    if (!dragged) { return; }
    var target = event.target.closest && event.target.closest("[data-task-list] .task");
    if (!target || listOf(target) !== listOf(dragged)) { return; }
    event.preventDefault();
    if (target === dragged) { return; }
    var rect = target.getBoundingClientRect();
    var after = event.clientY > rect.top + rect.height / 2;
    target.parentNode.insertBefore(dragged, after ? target.nextSibling : target);
  });

  document.addEventListener("drop", function (event) {
    if (dragged) { event.preventDefault(); }
  });

  document.addEventListener("dragend", function () {
    if (!dragged) { return; }
    var task = dragged;
    var list = listOf(task);
    var previous = idsAtStart;
    dragged = null;
    task.classList.remove("task-dragging");
    if (list && !sameIds(currentIds(list), previous)) {
      saveOrder(list, previous);
    }
  });

  document.addEventListener("keydown", function (event) {
    var handle = event.target.closest && event.target.closest("[data-drag-handle]");
    if (!handle || !event.altKey) { return; }
    var up = event.key === "ArrowUp";
    if (!up && event.key !== "ArrowDown") { return; }
    event.preventDefault();
    if (saving) { return; }

    var task = handle.closest(".task");
    var list = listOf(task);
    var sibling = up ? task.previousElementSibling : task.nextElementSibling;
    if (!sibling) { return; }

    var previous = currentIds(list);
    list.insertBefore(task, up ? sibling : sibling.nextSibling);
    handle.focus();

    var ids = currentIds(list);
    var status = document.querySelector("[data-reorder-status]");
    if (status) {
      status.textContent = "Tarea movida a la posición " + (ids.indexOf(task.dataset.taskId) + 1) + " de " + ids.length;
    }
    saveOrder(list, previous);
  });
})();
