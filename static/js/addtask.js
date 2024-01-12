document.getElementById("addTask").addEventListener("submit", function (event) {
    event.preventDefault(); // Prevent the default form submission
    
    var task = document.getElementById("task").value;
    var urgency = document.getElementById("urgency").value;
    var hours = parseFloat(document.getElementById("hours").value);
  
    // Capture the form data
    var data = {
      task: task,
      urgency: urgency,
      hours: hours,
      completed: false
    };
  
    // Create the AJAX request
    var xhr = new XMLHttpRequest();
    xhr.open("POST", "/tasks", true);
    xhr.setRequestHeader("Content-Type", "application/json");
  
    // Send the collected data as JSON
    xhr.send(JSON.stringify(data));
  
    xhr.onloadend = response => {
      if (xhr.status === 201) {
        // Clear the form
        document.getElementById("addTask").reset();
        alert("Task Added Successfully");
      } else {
        alert("Error! Please try again.");
        console.error(JSON.parse(response));
      }
    };
  });