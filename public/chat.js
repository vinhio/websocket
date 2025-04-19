window.onload = function () {
  const msg = document.getElementById("msg");
  const log = document.getElementById("log");
  const channelSelect = document.getElementById("channel");
  const currentChannel = document.getElementById("currentChannel");

  // The global WebSocket connection is initialized in auth.js
  // We'll use that connection for chat as well

  // Expose the WebSocket connection so it can be closed on logout
  window.closeWebSocketConnection = function() {
    if (window.globalConn) {
      window.globalConn.close();
      window.globalConn = null;
    }
  };

  // Create a chat message handler that will be called by the global message router
  window.chatMessageHandler = function(evt) {
    const messages = evt.data.split("\n");
    for (let i = 0; i < messages.length; i++) {
      try {
        // Try to parse the message as JSON
        const jsonData = JSON.parse(messages[i]);

        // Create a message display element
        const item = document.createElement("div");
        item.className = "mb-2 p-3 rounded-md border border-gray-100";

        // Check if this is an action-based message
        if (jsonData.action && jsonData.action.type) {
          // Handle different action types
          switch (jsonData.action.type) {
            case "list_channels":
              // Display the list of channels
              if (jsonData.action.data && jsonData.action.data.channels) {
                const channels = jsonData.action.data.channels;
                item.innerHTML = "<b>Available Channels:</b>";
                item.className += " bg-gray-50 text-gray-700";
                appendLog(item);

                // Update the dropdown list with the channels
                // Keep only the first option (General) and remove the rest
                while (channelSelect.options.length > 1) {
                  channelSelect.remove(1);
                }

                // Add all channels from the server to the dropdown
                for (let j = 0; j < channels.length; j++) {
                  // Skip the default/general channel as it's already in the dropdown
                  if (channels[j].id === "" || channels[j].id === "default" || 
                      channels[j].id.toLowerCase() === "general") {
                    continue;
                  }

                  // Check if the channel already exists in the dropdown
                  let exists = false;
                  for (let k = 0; k < channelSelect.options.length; k++) {
                    if (channelSelect.options[k].value === channels[j].id) {
                      exists = true;
                      break;
                    }
                  }

                  // If it doesn't exist, add it
                  if (!exists) {
                    const option = document.createElement("option");
                    option.value = channels[j].id;
                    option.text = channels[j].id;
                    channelSelect.add(option);
                  }
                }

                // Create a list of channels in the log
                for (let j = 0; j < channels.length; j++) {
                  const channelItem = document.createElement("div");
                  channelItem.className = "mb-2 p-3 rounded-md border border-gray-100 bg-gray-50 text-gray-700 flex items-center";
                  channelItem.innerHTML = `<span class="flex-1">- ${channels[j].id}</span>`;

                  // Add a button to switch to this channel
                  const switchBtn = document.createElement("button");
                  switchBtn.innerText = "Switch";
                  switchBtn.className = "bg-primary hover:bg-primary/90 text-white text-sm font-medium py-1 px-3 rounded-md transition-colors";
                  switchBtn.onclick = function() {
                    connectToChannel(channels[j].id);
                  };
                  channelItem.appendChild(switchBtn);

                  appendLog(channelItem);
                }
              }
              break;

            default:
              // For other action types, just display the action type
              item.innerHTML = `<b>Received action: ${jsonData.action.type}</b>`;
              item.className += " bg-gray-50 text-gray-700";
              appendLog(item);
          }
        } else if (jsonData.message) {
          // Legacy message format
          // Format the timestamp
          const timestamp = new Date(jsonData.message.timestamp);
          const timeStr = timestamp.toLocaleTimeString();

          // Get the sender ID
          const sender = jsonData.message.sender_id;

          // Get the message text
          let messageText = "";
          if (jsonData.message.type === "text" && jsonData.message.content) {
            messageText = jsonData.message.content.text;
          } else if (jsonData.message.content) {
            messageText = JSON.stringify(jsonData.message.content);
          }

          // Format the message with sender and timestamp
          item.innerHTML = `
            <div class="flex items-start">
              <div class="flex-1">
                <div class="flex items-center mb-1">
                  <span class="text-xs text-gray-500 mr-2">[${timeStr}]</span>
                  <span class="font-semibold text-gray-800">${sender}</span>
                </div>
                <div class="text-gray-700">${messageText}</div>
              </div>
            </div>
          `;

          // Add special styling for system messages
          if (sender === "system") {
            item.className += " bg-gray-50 text-gray-600 italic";
          } else {
            // Regular user message
            item.className += " hover:bg-gray-50";
          }

          appendLog(item);
        } else {
          // Unknown message format, display as JSON
          item.className += " font-mono text-xs p-2 bg-gray-50";
          item.innerText = JSON.stringify(jsonData, null, 2);
          appendLog(item);
        }
      } catch (e) {
        // If not valid JSON, display as plain text
        const item = document.createElement("div");
        item.className = "mb-2 p-3 rounded-md border border-gray-100 bg-gray-50 text-gray-700";
        item.innerText = messages[i];
        appendLog(item);
      }
    }
  };

  function appendLog(item) {
    const doScroll =
            log.scrollTop > log.scrollHeight - log.clientHeight - 1;
    log.appendChild(item);
    if (doScroll) {
      log.scrollTop = log.scrollHeight - log.clientHeight;
    }
  }

  document.getElementById("form").onsubmit = function () {
    if (!window.globalConn || window.globalConn.readyState !== WebSocket.OPEN) {
      return false;
    }
    if (!msg.value) {
      return false;
    }

    // Get current username from auth.js
    const username = getCurrentUsername();
    if (!username) {
      // Not logged in, don't send message
      return false;
    }

    // Create a JSON message using the new action-based structure
    const message = {
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
        id: channelSelect.value || "default",
        type: "group",
        participants: [],
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      },
      action: {
        type: "send_message",
        data: {
          message: {
            id: generateRandomId(),
            sender_id: username, // Use the authenticated username
            timestamp: new Date().toISOString(),
            type: "text",
            content: {
              text: msg.value
            },
            status: "sent",
            reactions: []
          }
        }
      }
    };

    // Send as JSON
    window.globalConn.send(JSON.stringify(message));
    msg.value = "";
    return false;
  };

  // Function to generate a random ID for messages
  function generateRandomId() {
    return Array.from(window.crypto.getRandomValues(new Uint8Array(16)))
      .map(b => b.toString(16).padStart(2, "0"))
      .join("");
  }

  // Function to connect to a specific channel
  // This function has been improved to maintain the WebSocket connection when switching channels
  // Instead of closing and reopening the connection, it now sends a channel switch command
  // to the server, which allows the server to move the client to a different channel
  // without disconnecting the WebSocket connection.
  function connectToChannel(channel) {
    // Update current channel display
    currentChannel.textContent = channel || "General";

    // We're using the global WebSocket connection established in auth.js
    if (!window.globalConn || window.globalConn.readyState !== WebSocket.OPEN) {
      // If the global connection doesn't exist or isn't open, we need to wait
      // It should be initialized in auth.js, but we'll add a fallback just in case
      const item = document.createElement("div");
      item.className = "mb-2 p-3 rounded-md bg-warning/10 text-warning border border-warning/20 text-center";
      item.innerHTML = "<b>Waiting for connection to be established...</b>";
      appendLog(item);

      // Try to initialize the global connection if it doesn't exist
      if (!window.globalConn && typeof initializeGlobalWebSocket === 'function') {
        initializeGlobalWebSocket();
      }

      // Add a connection status message when the connection is established
      if (window.globalConn) {
        window.globalConn.addEventListener('open', function() {
          const connectedItem = document.createElement("div");
          connectedItem.className = "mb-2 p-3 rounded-md bg-success/10 text-success border border-success/20 text-center";
          connectedItem.innerHTML = "<b>Connected to channel: " + (channel || "General") + "</b>";
          appendLog(connectedItem);

          // Send the channel switch command once connected
          sendChannelSwitchCommand(channel);
        });
      }
    } else {
      // If the global connection exists and is open, send a channel switch command
      sendChannelSwitchCommand(channel);
    }
  }

  // Helper function to send a channel switch command
  function sendChannelSwitchCommand(channel) {
    // Create a channel switch message using the action-based structure
    const switchMessage = {
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
        id: channel || "default",
        type: "group",
        participants: [],
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      },
      action: {
        type: "switch_channel",
        data: {
          channel_id: channel || "default"
        }
      }
    };

    // Send as JSON
    window.globalConn.send(JSON.stringify(switchMessage));

    const item = document.createElement("div");
    item.className = "mb-2 p-3 rounded-md bg-primary/10 text-primary border border-primary/20 text-center";
    item.innerHTML = "<b>Switching to channel: " + (channel || "General") + "</b>";
    appendLog(item);
  }

  // Handle channel selection change
  channelSelect.addEventListener("change", function() {
    const selectedChannel = channelSelect.value;
    connectToChannel(selectedChannel);
  });

  // Handle new channel button click
  document.getElementById("newChannelBtn").addEventListener("click", function() {
    const channelName = prompt("Enter new channel name:");
    if (channelName) {
      // Check if the channel already exists
      let channelExists = false;
      for (let i = 0; i < channelSelect.options.length; i++) {
        if (channelSelect.options[i].value === channelName) {
          channelExists = true;
          channelSelect.selectedIndex = i;
          break;
        }
      }

      // If the channel doesn't exist, add it
      if (!channelExists) {
        const option = document.createElement("option");
        option.value = channelName;
        option.text = channelName;
        channelSelect.add(option);
        channelSelect.value = channelName;

        // Send a create channel request using the action-based structure
        if (window.globalConn && window.globalConn.readyState === WebSocket.OPEN) {
          const createChannelMessage = {
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
              id: channelSelect.value || "default",
              type: "group",
              participants: [],
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString()
            },
            action: {
              type: "create_channel",
              data: {
                channel: {
                  id: channelName,
                  type: "group",
                  participants: [],
                  created_at: new Date().toISOString(),
                  updated_at: new Date().toISOString()
                }
              }
            }
          };

          // Send as JSON
          window.globalConn.send(JSON.stringify(createChannelMessage));

          const item = document.createElement("div");
          item.className = "mb-2 p-3 rounded-md bg-success/10 text-success border border-success/20 text-center";
          item.innerHTML = "<b>Creating new channel: " + channelName + "</b>";
          appendLog(item);
        }
      }

      // Connect to the new channel
      connectToChannel(channelName);
    }
  });

  // Handle list channels button click
  document.getElementById("listChannelsBtn").addEventListener("click", function() {
    if (!window.globalConn || window.globalConn.readyState !== WebSocket.OPEN) {
      const item = document.createElement("div");
      item.className = "mb-2 p-3 rounded-md bg-warning/10 text-warning border border-warning/20 text-center";
      item.innerHTML = "<b>Not connected to server.</b>";
      appendLog(item);
      return;
    }

    // Create a list channels request using the action-based structure
    const listChannelsMessage = {
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
        id: channelSelect.value || "default",
        type: "group",
        participants: [],
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      },
      action: {
        type: "list_channels",
        data: {}
      }
    };

    // Send as JSON
    window.globalConn.send(JSON.stringify(listChannelsMessage));

    const item = document.createElement("div");
    item.className = "mb-2 p-3 rounded-md bg-info/10 text-info border border-info/20 text-center";
    item.innerHTML = "<b>Requesting channel list...</b>";
    appendLog(item);
  });

  // Function to initialize chat after successful login
  // This will be called after successful login
  // The WebSocket connection is already established in auth.js
  function initializeWebSocket() {
    if (window["WebSocket"]) {
      // Get channel from URL if present
      const urlParams = new URLSearchParams(window.location.search);
      const channelParam = urlParams.get('channel');

      // Set the select value if channel is in URL
      if (channelParam) {
        // Check if the channel exists in the dropdown
        for (let i = 0; i < channelSelect.options.length; i++) {
          if (channelSelect.options[i].value === channelParam) {
            channelSelect.selectedIndex = i;
            break;
          }
        }
        // If not found, add it
        if (channelSelect.value !== channelParam) {
          const option = document.createElement("option");
          option.value = channelParam;
          option.text = channelParam;
          channelSelect.add(option);
          channelSelect.value = channelParam;
        }
      }

      // Connect to the selected channel using the existing global connection
      connectToChannel(channelSelect.value);

      // Add a welcome message
      const item = document.createElement("div");
      item.className = "mb-2 p-3 rounded-md bg-success/10 text-success border border-success/20 text-center";
      item.innerHTML = "<b>Welcome to the chat! You are now connected.</b>";
      appendLog(item);
    } else {
      const item = document.createElement("div");
      item.className = "mb-2 p-3 rounded-md bg-danger/10 text-danger border border-danger/20 text-center";
      item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
      appendLog(item);
    }
  }

  // Expose the initializeWebSocket function globally
  window.initializeWebSocket = initializeWebSocket;
};
