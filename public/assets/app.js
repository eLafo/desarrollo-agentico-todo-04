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
  // y _method=PUT, que Buffalo convierte en PUT /tasks/{id}/toggle. La acción redirige a
  // la lista y fetch sigue la redirección: el recuento de pendientes de esa página, que
  // pinta el servidor con el cambio ya guardado y con el número en negrita, sustituye al
  // de la cabecera.
  fetch(form.action, {
    method: "POST",
    body: new URLSearchParams(new FormData(form)),
  })
    .then(function (response) {
      return response.ok ? response.text() : "";
    })
    .then(function (html) {
      var current = document.querySelector(".list-count");
      var saved = new DOMParser().parseFromString(html, "text/html").querySelector(".list-count");
      if (current && saved) {
        current.replaceWith(saved);
      }
    });
});
