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

  // Guarda el cambio con la misma petición que el formulario: POST con el token CSRF
  // y _method=PUT, que Buffalo convierte en PUT /tasks/{id}/toggle.
  fetch(form.action, {
    method: "POST",
    body: new URLSearchParams(new FormData(form)),
  });
});

// Reordenar: el ratón arrastra las tareas y los botones «Subir» y «Bajar» las mueven un
// puesto. La página cambia al momento y la nueva posición se guarda con el formulario de
// mover de la tarea.
(function () {
  var list = document.querySelector(".task-list");
  if (!list) {
    return;
  }

  var saving = Promise.resolve();

  function indexOf(task) {
    return Array.prototype.indexOf.call(list.children, task);
  }

  // Guarda el lugar que ocupa la tarea en la página con la misma petición que el formulario:
  // POST con el token CSRF y _method=PUT, que Buffalo convierte en PUT /tasks/{id}/move.
  // Los envíos van en cadena para que lleguen en orden; un fallo no corta la cadena.
  function savePosition(task) {
    var form = task.querySelector("[data-task-move]");
    var body = new URLSearchParams(new FormData(form));
    body.set("position", String(indexOf(task)));
    saving = saving
      .then(function () {
        return fetch(form.action, { method: "POST", body: body });
      })
      .catch(function () {});
  }

  // «Subir» y «Bajar»: se mueve la tarea vecina y no la propia, para que el foco siga en el botón.
  document.addEventListener("submit", function (event) {
    var form = event.target;
    if (!form.matches("[data-task-move]")) {
      return;
    }
    event.preventDefault();

    var task = form.closest(".task");
    var direction = event.submitter && event.submitter.dataset.direction;
    if (direction === "up" && task.previousElementSibling) {
      task.after(task.previousElementSibling);
    } else if (direction === "down" && task.nextElementSibling) {
      task.before(task.nextElementSibling);
    } else {
      return;
    }
    savePosition(task);
  });

  // Arrastrar y soltar con el ratón. La lista se reordena mientras se arrastra; al soltar
  // dentro de ella se guarda, y si se cancela la tarea vuelve a su sitio.
  var dragged = null;
  var dropped = false;
  var startIndex = -1;
  var startNext = null;

  list.querySelectorAll(".task").forEach(function (task) {
    task.draggable = true;
  });

  list.addEventListener("dragstart", function (event) {
    if (!(event.target instanceof Element) || !event.target.matches(".task")) {
      return; // los enlaces de la tarea se arrastran como siempre
    }
    dragged = event.target;
    dropped = false;
    startIndex = indexOf(dragged);
    startNext = dragged.nextElementSibling;
    event.dataTransfer.effectAllowed = "move";
    // Firefox no empieza a arrastrar sin datos.
    event.dataTransfer.setData("text/plain", dragged.querySelector(".task-title").textContent);
    dragged.classList.add("task-dragging");
  });

  list.addEventListener("dragover", function (event) {
    if (!dragged) {
      return;
    }
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";

    // En Firefox el destino puede ser un nodo de texto.
    var node = event.target instanceof Element ? event.target : event.target.parentElement;
    var over = node && node.closest(".task");
    if (!over || over === dragged) {
      return;
    }
    var box = over.getBoundingClientRect();
    list.insertBefore(dragged, event.clientY < box.top + box.height / 2 ? over : over.nextElementSibling);
  });

  list.addEventListener("drop", function (event) {
    if (!dragged) {
      return;
    }
    event.preventDefault();
    dropped = true;
    if (indexOf(dragged) !== startIndex) {
      savePosition(dragged);
    }
  });

  list.addEventListener("dragend", function () {
    if (!dragged) {
      return;
    }
    dragged.classList.remove("task-dragging");
    if (!dropped) {
      list.insertBefore(dragged, startNext);
    }
    dragged = null;
  });
})();
