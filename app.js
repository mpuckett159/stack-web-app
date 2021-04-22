new Vue({
    el: '#app',

    data: {
        ws: null, // Our websocket
        tableId: null, // meeting ID
        stackContent: '', // Current speaker queue stack
        name: null,
        joined: false,
        meetingUrl: null
    },

    created: function() {
        let urlParams = new URLSearchParams(window.location.search);
        if (urlParams.has('meeting_id')) {
            this.tableId = urlParams.get('meeting_id');
        };
        if (window.location.protocol == "http:"){
            this.wsProtocol = "ws:"
        } else {
            this.wsProtocol = "wss:"
        };
    },

    methods: {
        sendon: function () {
            if (this.newMsg != '') {
                this.ws.send(
                    JSON.stringify({
                        tableId: this.tableId,
                        action: "on",
                        name: this.name
                    }
                ));
            }
        },

        sendoff: function () {
            if (this.newMsg != '') {
                this.ws.send(
                    JSON.stringify({
                        tableId: this.tableId,
                        action: "off",
                        name: this.name
                    }
                ));
            }
        },

        join: function () {
            // Validate inputs
            if (!this.tableId && !this.name) {
                Materialize.toast('You must set a meeting ID and name to join a meeting.', 2000);
                return
            }
            if (!this.tableId) {
                Materialize.toast('You must set a meeting ID', 2000);
                return
            }
            if(!this.name) {
                Materialize.toast('You must set a name', 2000);
                return
            }

            // Store self for later reference
            var self = this;

            // Join an on-going meeting by ID and update session values
            this.meetingUrl = window.location.host + '/?meeting_id=' + this.tableId;
            this.ws = new WebSocket(this.wsProtocol + '//' + window.location.host + '/ws?meeting_id=' + this.tableId);
            this.joined = true;

            // Set up event listeners to handle incoming/outgoing messages and open/close actions
            this.ws.addEventListener('message', function(e) { messageListener(self, e) });
        },

        create: async function (submitEvent) {
            // Validate inputs
            if(!this.name) {
                Materialize.toast('You must set a name', 2000);
                return
            }

            // Store self for later reference
            var self = this;

            // Execute fetch to create a new WebSocket hub to connect to and store info back
            const requestOptions = {
                method: "POST",
                headers: { "Accept": "application/json" }
            };
            await fetch(window.location.protocol + "//" + window.location.host + "/ws", requestOptions)
              .then(response => response.json())
              .then(data => self.tableId = data.meetingId);

            // Set up new WebSocket to be used with the required meeting ID and update session values
            this.meetingUrl = window.location.host + '/?meeting_id=' + this.tableId;
            this.ws = new WebSocket(this.wsProtocol + '//' + window.location.host + '/ws?meeting_id=' + this.tableId);
            this.joined = true;

            // Set up event listeners to handle incoming/outgoing messages and open/close actions
            this.ws.addEventListener('message', function(e) { messageListener(self, e) });
        }
    }
});

function messageListener(self, event) {
    self.stackContent = ""
    var data = $.parseJSON(event.data);
    if (data) {
        for(var speaker in data) {
            var id = data[speaker].speakerId.split("-")
            self.stackContent += '<div class="chip">'
                + data[speaker].name + "<br/>" + id[id.length - 1]
            + '</div><br/>';
        };
    }
}

function copyMeetingUrl() {
    /* Get the text field */
    var copyText = document.getElementById("meetingUrl");
  
    /* Select the text field */
    copyText.select();
    copyText.setSelectionRange(0, 99999); /* For mobile devices */
  
    /* Copy the text inside the text field */
    document.execCommand("copy");
  
    /* Alert the copied text */
    Materialize.toast('Meeting URL copied.', 2000);
}