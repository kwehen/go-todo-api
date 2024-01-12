//Extract ID from URL and set it as the value of the hidden input field
var path = window.location.pathname; // This will return "/task/1"
var segments = path.split('/'); // This will split the path into segments: ["", "task", "1"]
var taskId = segments[2];
//Add event listener to the form to submit the form when the button is clicked
document.getElementById('completeThisTask').addEventListener('submit', function(e) {
    e.preventDefault();

    var url = '/completeTask/' + taskId;
    var xhr = new XMLHttpRequest();
    xhr.open('POST', url, true);
    xhr.setRequestHeader("Content-Type", "application/json");

    // Create a FormData object
    var formData = new FormData();
    // Append the task ID to the form data
    formData.append('id', taskId);

    // Send the form data in the POST request
    xhr.send(JSON.stringify(Object.fromEntries(formData)));

    xhr.onloadend = response => {
        if (xhr.status == 200) {
            alert('Task Completed!');
            location.reload();
        } else {
            alert('Error completing task');
            console.error(JSON.parse(xhr.response));
        }
    };
});

// document.addEventListener("DOMContentLoaded", function() {
//     document.querySelectorAll('.task-row').forEach(row => {
//         row.addEventListener('click', function() {
//             window.location.href = '/tasks/' + this.dataset.id;
//         });
//     });
// });