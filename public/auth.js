// auth.js - Handles user authentication

// Global WebSocket connection that will be used for both authentication and chat
window.globalConn = null;

// Store the current user and authentication state
let currentUser = null;
let authCallback = null;
let authMessageHandler = null;

// Check if user is logged in
function isLoggedIn() {
  return currentUser !== null;
}

// Get current username
function getCurrentUsername() {
  return currentUser;
}

// Initialize the global WebSocket connection
function initializeGlobalWebSocket() {
  if (window.globalConn === null) {
    const wsUrl = "ws://" + document.location.host + "/ws";
    window.globalConn = new ReconnectingWebSocket(wsUrl);

    console.log("Global WebSocket connection initialized");

    // Set up a message handler that will route messages to the appropriate handler
    window.globalConn.onmessage = function(evt) {
      try {
        const response = JSON.parse(evt.data);

        // If we have an auth message handler and this is an auth response, route to it
        if (authMessageHandler && response.action && response.action.type === "user_auth") {
          authMessageHandler(evt);
        } 
        // Otherwise, if we have a chat message handler, route to it
        else if (window.chatMessageHandler) {
          window.chatMessageHandler(evt);
        }
      } catch (e) {
        console.error("Error parsing WebSocket message:", e);
      }
    };

    window.globalConn.onerror = function(evt) {
      console.error("WebSocket error:", evt);

      // If we have an auth callback and we're not logged in yet, report the error
      if (authCallback && !isLoggedIn()) {
        authCallback(false, "Connection error");
        authCallback = null;
      }
    };
  }

  return window.globalConn;
}

// Initialize the global WebSocket connection when the page loads
document.addEventListener('DOMContentLoaded', function() {
  initializeGlobalWebSocket();
});

// Login function - uses the global WebSocket connection for authentication
function login(username, password, callback) {
  // Store the callback to be called when we get a response
  authCallback = callback;

  // Make sure we have a global connection
  if (!window.globalConn) {
    initializeGlobalWebSocket();
  }

  // Set up the auth message handler
  authMessageHandler = function(evt) {
    try {
      const response = JSON.parse(evt.data);

      // Check if this is an authentication response
      if (response.action && response.action.type === "user_auth") {
        const authData = response.action.data;

        if (authData.success) {
          // Authentication successful
          currentUser = authData.username;

          // Call the callback with success
          if (authCallback) {
            authCallback(true, authData.message);
            authCallback = null;
          }

          // We no longer need the auth message handler
          authMessageHandler = null;
        } else {
          // Authentication failed
          currentUser = null;

          // Call the callback with failure
          if (authCallback) {
            authCallback(false, authData.message);
            authCallback = null;
          }
        }
      }
    } catch (e) {
      console.error("Error parsing auth response:", e);

      // Call the callback with failure
      if (authCallback) {
        authCallback(false, "Error processing server response");
        authCallback = null;
      }
    }
  };

  // Wait for the connection to be open before sending the auth request
  if (window.globalConn.readyState === WebSocket.OPEN) {
    sendAuthRequest(username, password);
  } else {
    window.globalConn.addEventListener('open', function() {
      sendAuthRequest(username, password);
    });
  }

  // Return false as the actual result will come asynchronously
  return false;
}

// Helper function to send the authentication request
function sendAuthRequest(username, password) {
  console.log("Sending authentication request");

  // Create authentication request
  const authRequest = {
    metadata: {
      version: "1.0",
      timestamp: new Date().toISOString(),
      server_node: {
        id: "browser-client",
        region: "local",
        load: 0.0
      }
    },
    channel: {
      id: "auth",
      type: "system",
      participants: [],
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    },
    action: {
      type: "user_auth",
      data: {
        username: username,
        password: password
      }
    }
  };

  // Send authentication request
  window.globalConn.send(JSON.stringify(authRequest));
}

// Logout function
function logout() {
  currentUser = null;
  authMessageHandler = null;

  // Note: We don't close the global WebSocket connection here
  // It will be closed by the closeWebSocketConnection function in chat.js if needed
}

// Initialize authentication
document.addEventListener('DOMContentLoaded', function() {
  const loginForm = document.getElementById('login-form');
  const loginError = document.getElementById('login-error');
  const loginContainer = document.getElementById('login-container');
  const chatContainer = document.getElementById('chat-container');
  const currentUserSpan = document.getElementById('currentUser');
  const logoutBtn = document.getElementById('logoutBtn');

  // Handle login form submission
  loginForm.addEventListener('submit', function(e) {
    e.preventDefault();

    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    // Disable form while logging in
    const submitButton = loginForm.querySelector('button[type="submit"]');
    submitButton.disabled = true;
    submitButton.textContent = 'Logging in...';
    loginError.classList.add('hidden');

    // Call the asynchronous login function with a callback
    login(username, password, function(success, message) {
      // Re-enable form
      submitButton.disabled = false;
      submitButton.textContent = 'Login';

      if (success) {
        // Hide login form and show chat
        loginContainer.classList.add('hidden');
        chatContainer.classList.remove('hidden');

        // Display current user
        currentUserSpan.textContent = username;

        // Initialize WebSocket connection
        if (window.initializeWebSocket) {
          window.initializeWebSocket();
        }

        // Focus on message input
        document.getElementById('msg').focus();
      } else {
        // Show error message
        loginError.textContent = message || 'Authentication failed';
        loginError.classList.remove('hidden');
      }
    });
  });

  // Handle logout button click
  logoutBtn.addEventListener('click', function() {
    logout();

    // Close WebSocket connection
    if (window.closeWebSocketConnection) {
      window.closeWebSocketConnection();
    }

    // Hide chat and show login form
    chatContainer.classList.add('hidden');
    loginContainer.classList.remove('hidden');

    // Clear login form
    document.getElementById('username').value = '';
    document.getElementById('password').value = '';
    loginError.classList.add('hidden');

    // Clear chat log
    document.getElementById('log').innerHTML = '';
  });
});
