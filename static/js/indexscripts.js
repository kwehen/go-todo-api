document.getElementById("completed").addEventListener("submit", function (event) {
  event.preventDefault(); // Prevent the default submission

  var idValue = document.getElementById("completedId").value;

  // Create the AJAX request
  var xhr = new XMLHttpRequest();
  xhr.open("GET", "/completed/" + encodeURIComponent(idValue), true);
  xhr.setRequestHeader("Content-Type", "application/json");
  
  // Send the collected data as JSON
  xhr.send();

  xhr.onloadend = response => {
    if (xhr.status === 200) {
      // Clear the form
      document.getElementById("completed").reset();
      alert("Task Completed and Moved to Completed Table");
    } else {
      alert("Error! Please try again.");
      console.error(JSON.parse(response));
    }
  };
});
