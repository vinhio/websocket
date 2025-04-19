// Function to get query parameters from the URL
function getQueryParams(qs) {
  qs = qs.split('+').join(' ');

  let params = {},
          tokens,
          re = /[?&]?([^=]+)=([^&]*)/g;

  while (tokens = re.exec(qs)) {
    params[decodeURIComponent(tokens[1])] = decodeURIComponent(tokens[2]);
  }

  return params;
}

// Function to generate a random string of specified length
function generateRandomString(length) {
  const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  let result = '';
  for (let i = 0; i < length; i++) {
    const randomIndex = Math.floor(Math.random() * characters.length);
    result += characters.charAt(randomIndex);
  }
  return result;
}

// Check if the URL contains the 'auto' query parameter
// If 'auto' is true, automatically send messages every 10 seconds
document.addEventListener('DOMContentLoaded', function() {
  const query = getQueryParams(document.location.search);
  if (query.auto === 'true') {
    const button = document.querySelector('input[type="submit"]');

    setInterval(function() {
      const msg = document.getElementById("msg");
      msg.value = generateRandomString(50);

      button.click();
    }, 30000);
  }
});