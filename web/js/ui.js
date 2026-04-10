export function showStatus(element, message, type = "info") {

    const form = element.closest("form");

    let existing = form.querySelector(".status-message");
    if (existing) existing.remove();

    const msg = document.createElement("div");
    msg.textContent = message;
    msg.classList.add("status-message");

    if (type === "error") msg.style.color = "red";
    if (type === "success") msg.style.color = "green";
    if (type === "info") msg.style.color = "gray";

    form.prepend(msg);
}