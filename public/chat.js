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

        // Create a message display element with Flowbite styling
        const item = document.createElement("div");
        item.className = "mb-3 p-4 rounded-lg border border-gray-200 shadow-sm";

        // Check if this is an action-based message
        if (jsonData.action && jsonData.action.type) {
          // Handle different action types
          switch (jsonData.action.type) {
            case "list_channels":
              // Display the list of channels
              if (jsonData.action.data && jsonData.action.data.channels) {
                const channels = jsonData.action.data.channels;
                item.innerHTML = `
                  <div class="flex items-center">
                    <svg class="w-5 h-5 text-primary mr-2" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 20 18">
                      <path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 5h2a1 1 0 0 1 1 1v7a1 1 0 0 1-1 1h-2v3l-4-3H8m4-13H2a1 1 0 0 0-1 1v7a1 1 0 0 0 1 1h2v3l4-3h4a1 1 0 0 0 1-1V2a1 1 0 0 0-1-1Z"/>
                    </svg>
                    <span class="font-medium text-lg text-gray-800">Available Channels</span>
                  </div>
                `;
                item.className += " bg-blue-50 border-blue-200 text-gray-700";
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
                  channelItem.className = "mb-2 p-3 rounded-lg border border-gray-200 bg-gray-50 text-gray-700 flex items-center";

                  // Create channel name with icon
                  const channelNameContainer = document.createElement("div");
                  channelNameContainer.className = "flex items-center flex-1";

                  const channelIcon = document.createElement("span");
                  channelIcon.className = "inline-flex items-center justify-center w-8 h-8 me-2 text-sm font-semibold text-primary bg-blue-100 rounded-full";
                  channelIcon.innerHTML = channels[j].id.charAt(0).toUpperCase();

                  const channelName = document.createElement("span");
                  channelName.className = "font-medium text-gray-800";
                  channelName.textContent = channels[j].id || "General";

                  channelNameContainer.appendChild(channelIcon);
                  channelNameContainer.appendChild(channelName);
                  channelItem.appendChild(channelNameContainer);

                  // Add a button to switch to this channel
                  const switchBtn = document.createElement("button");
                  switchBtn.className = "text-white bg-primary hover:bg-primary-800 focus:ring-4 focus:ring-primary-300 font-medium rounded-lg text-xs px-3 py-1.5 inline-flex items-center";
                  switchBtn.innerHTML = `
                    <svg class="w-3 h-3 me-1.5" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 18 18">
                      <path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 1v16M1 9h16"/>
                    </svg>
                    Join
                  `;
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
          if (sender === "system") {
            // System message with special styling
            item.innerHTML = `
              <div class="flex items-start">
                <div class="flex-shrink-0">
                  <span class="inline-flex items-center justify-center h-8 w-8 rounded-full bg-gray-200">
                    <svg class="w-4 h-4 text-gray-500" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 18">
                      <path d="M18 0H2a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h3.546l3.2 3.659a1 1 0 0 0 1.506 0L13.454 14H18a2 2 0 0 0 2-2V2a2 2 0 0 0-2-2Zm-8 10H5a1 1 0 0 1 0-2h5a1 1 0 1 1 0 2Zm5-4H5a1 1 0 0 1 0-2h10a1 1 0 1 1 0 2Z"/>
                    </svg>
                  </span>
                </div>
                <div class="ms-3 flex-1">
                  <div class="flex items-center mb-1">
                    <span class="text-xs font-medium text-gray-500 mr-2">${timeStr}</span>
                    <span class="text-sm font-semibold text-gray-700">System</span>
                  </div>
                  <div class="text-sm text-gray-600 italic">${messageText}</div>
                </div>
              </div>
            `;
            item.className += " bg-gray-50 border-gray-200";
          } else {
            // Regular user message
            // Generate a consistent color based on the sender's name
            const colors = ['blue', 'green', 'purple', 'pink', 'yellow', 'red', 'indigo'];
            const colorIndex = sender.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0) % colors.length;
            const color = colors[colorIndex];

            item.innerHTML = `
              <div class="flex items-start">
                <div class="flex-shrink-0">
                  <span class="inline-flex items-center justify-center h-8 w-8 rounded-full bg-${color}-100 text-${color}-700">
                    ${sender.charAt(0).toUpperCase()}
                  </span>
                </div>
                <div class="ms-3 flex-1">
                  <div class="flex items-center mb-1">
                    <span class="text-xs font-medium text-gray-500 mr-2">${timeStr}</span>
                    <span class="text-sm font-semibold text-gray-800">${sender}</span>
                  </div>
                  <div class="text-sm text-gray-700">${messageText}</div>
                </div>
              </div>
            `;
            item.className += " hover:bg-gray-50 border-gray-200";
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
      item.className = "flex p-4 mb-4 text-sm text-yellow-800 border border-yellow-300 rounded-lg bg-yellow-50";
      item.innerHTML = `
        <svg class="flex-shrink-0 inline w-4 h-4 me-3 mt-[2px]" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20">
          <path d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5ZM10 15a1 1 0 1 1 0-2 1 1 0 0 1 0 2Zm1-4a1 1 0 0 1-2 0V6a1 1 0 0 1 2 0v5Z"/>
        </svg>
        <span class="sr-only">Warning</span>
        <div>
          <span class="font-medium">Waiting for connection to be established...</span>
        </div>
      `;
      appendLog(item);

      // Try to initialize the global connection if it doesn't exist
      if (!window.globalConn && typeof initializeGlobalWebSocket === 'function') {
        initializeGlobalWebSocket();
      }

      // Add a connection status message when the connection is established
      if (window.globalConn) {
        window.globalConn.addEventListener('open', function() {
          const connectedItem = document.createElement("div");
          connectedItem.className = "flex p-4 mb-4 text-sm text-green-800 border border-green-300 rounded-lg bg-green-50";
          connectedItem.innerHTML = `
            <svg class="flex-shrink-0 inline w-4 h-4 me-3 mt-[2px]" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20">
              <path d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5Zm3.707 8.207-4 4a1 1 0 0 1-1.414 0l-2-2a1 1 0 0 1 1.414-1.414L9 10.586l3.293-3.293a1 1 0 0 1 1.414 1.414Z"/>
            </svg>
            <span class="sr-only">Success</span>
            <div>
              <span class="font-medium">Connected to channel: ${channel || "General"}</span>
            </div>
          `;
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
    item.className = "flex p-4 mb-4 text-sm text-blue-800 border border-blue-300 rounded-lg bg-blue-50";
    item.innerHTML = `
      <svg class="flex-shrink-0 inline w-4 h-4 me-3 mt-[2px]" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20">
        <path d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5Zm1 13a1 1 0 1 1-2 0 1 1 0 0 1 2 0Zm-.25-6.25a.75.75 0 0 0-1.5 0v3.5a.75.75 0 0 0 1.5 0v-3.5Z"/>
      </svg>
      <span class="sr-only">Info</span>
      <div>
        <span class="font-medium">Switching to channel: ${channel || "General"}</span>
      </div>
    `;
    appendLog(item);
  }

  // Handle channel selection change
  channelSelect.addEventListener("change", function() {
    const selectedChannel = channelSelect.value;
    connectToChannel(selectedChannel);
  });

  // Handle new channel button click - using data-modal-target attribute in HTML
  // The button already has data-modal-target and data-modal-toggle attributes in the HTML

  // Handle new channel form submission
  document.getElementById("newChannelForm").addEventListener("submit", function(e) {
    e.preventDefault();

    // Get the channel name from the form
    const channelNameInput = document.getElementById("channelName");
    const channelName = channelNameInput.value.trim();

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
          item.className = "flex p-4 mb-4 text-sm text-green-800 border border-green-300 rounded-lg bg-green-50";
          item.innerHTML = `
            <svg class="flex-shrink-0 inline w-4 h-4 me-3 mt-[2px]" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20">
              <path d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5Zm3.707 8.207-4 4a1 1 0 0 1-1.414 0l-2-2a1 1 0 0 1 1.414-1.414L9 10.586l3.293-3.293a1 1 0 0 1 1.414 1.414Z"/>
            </svg>
            <span class="sr-only">Success</span>
            <div>
              <span class="font-medium">Creating new channel: ${channelName}</span>
            </div>
          `;
          appendLog(item);
        }
      }

      // Connect to the new channel
      connectToChannel(channelName);

      // Reset the form and hide the modal
      channelNameInput.value = '';

      // Hide the modal using Flowbite's data attribute
      const modalHideButton = document.querySelector('[data-modal-hide="newChannelModal"]');
      if (modalHideButton) {
        modalHideButton.click();
      }
    }
  });

  // Handle list channels button click
  document.getElementById("listChannelsBtn").addEventListener("click", function() {
    if (!window.globalConn || window.globalConn.readyState !== WebSocket.OPEN) {
      const item = document.createElement("div");
      item.className = "flex p-4 mb-4 text-sm text-red-800 border border-red-300 rounded-lg bg-red-50";
      item.innerHTML = `
        <svg class="flex-shrink-0 inline w-4 h-4 me-3 mt-[2px]" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20">
          <path d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5Zm3.5 13.5a.75.75 0 1 1-1.5 0 .75.75 0 0 1 1.5 0Zm-.75-5.5a.75.75 0 0 0-1.5 0v3a.75.75 0 0 0 1.5 0v-3Z"/>
        </svg>
        <span class="sr-only">Error</span>
        <div>
          <span class="font-medium">Not connected to server.</span>
        </div>
      `;
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
    item.className = "flex p-4 mb-4 text-sm text-blue-800 border border-blue-300 rounded-lg bg-blue-50";
    item.innerHTML = `
      <svg class="flex-shrink-0 inline w-4 h-4 me-3 mt-[2px]" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20">
        <path d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5Zm1 13a1 1 0 1 1-2 0 1 1 0 0 1 2 0Zm-.25-6.25a.75.75 0 0 0-1.5 0v3.5a.75.75 0 0 0 1.5 0v-3.5Z"/>
      </svg>
      <span class="sr-only">Info</span>
      <div>
        <span class="font-medium">Requesting channel list...</span>
      </div>
    `;
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
      item.className = "flex p-4 mb-4 text-sm text-green-800 border border-green-300 rounded-lg bg-green-50";
      item.innerHTML = `
        <svg class="flex-shrink-0 inline w-4 h-4 me-3 mt-[2px]" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20">
          <path d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5Zm3.707 8.207-4 4a1 1 0 0 1-1.414 0l-2-2a1 1 0 0 1 1.414-1.414L9 10.586l3.293-3.293a1 1 0 0 1 1.414 1.414Z"/>
        </svg>
        <span class="sr-only">Success</span>
        <div>
          <span class="font-medium">Welcome to the chat! You are now connected.</span>
        </div>
      `;
      appendLog(item);
    } else {
      const item = document.createElement("div");
      item.className = "flex p-4 mb-4 text-sm text-red-800 border border-red-300 rounded-lg bg-red-50";
      item.innerHTML = `
        <svg class="flex-shrink-0 inline w-4 h-4 me-3 mt-[2px]" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20">
          <path d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5Zm3.5 13.5a.75.75 0 1 1-1.5 0 .75.75 0 0 1 1.5 0Zm-.75-5.5a.75.75 0 0 0-1.5 0v3a.75.75 0 0 0 1.5 0v-3Z"/>
        </svg>
        <span class="sr-only">Error</span>
        <div>
          <span class="font-medium">Your browser does not support WebSockets.</span>
        </div>
      `;
      appendLog(item);
    }
  }

  // Expose the initializeWebSocket function globally
  window.initializeWebSocket = initializeWebSocket;
};
