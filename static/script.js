document.getElementById('waitlist-form').addEventListener('submit', function(event) {
    event.preventDefault();
    const email = document.getElementById('email').value;
    // Here you can add your logic to handle the email submission, e.g., send it to a server
    document.getElementById('waitlist-message').textContent = `Thank you! We'll notify you at ${email} when we launch.`;
    document.getElementById('email').value = ''; // Clear the input field
});