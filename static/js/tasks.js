document.addEventListener("DOMContentLoaded", function() {
    document.querySelectorAll('.task-row').forEach(row => {
        row.addEventListener('click', function() {
            window.location.href = '/tasks/' + this.dataset.id;
        });
    });
});