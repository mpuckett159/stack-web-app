new Vue({
    el: '#app',

    data: {
        ws: null, // Our websocket
        tableId: null, // meeting ID
        stackContent: '', // Current speaker queue stack
        name: null,
        joined: false
    },

    created: function() {
        var self = this;
        let urlParams = new URLSearchParams(window.location.search);
        if (urlParams.has('meeting_id')) {
            this.tableId = urlParams.get('meeting_id');

            this.ws = new WebSocket('ws://' + window.location.host + '/ws?meeting_id=' + this.tableId);
            this.ws.addEventListener('message', function(e) {
                var msg = JSON.parse(e.data);
                self.stackContent += '<div class="chip">'
                        + emojione.toImage(msg.name)
                    + '</div><br/>';

                var element = document.getElementById('current-stack');
                element.scrollTop = element.scrollHeight; // Auto scroll to the bottom
            });
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

            // Join an on-going meeting by ID and update session values
            this.ws = new WebSocket('ws://' + window.location.host + '/ws?meeting_id=' + this.tableId);
            this.joined = true;
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
            await fetch("http://" + window.location.host + "/ws", requestOptions)
              .then(response => response.json())
              .then(data => self.tableId = data.meetingId);

            // Set up new WebSocket to be used with the required meeting ID and update session values
            this.ws = new WebSocket('ws://' + window.location.host + '/ws?meeting_id=' + this.tableId);
            this.joined = true;

            // Set up event listeners to handle incoming/outgoing messages and open/close actions
            this.ws.addEventListener('message', function(e) {
                self.stackContent = ""
                var data = $.parseJSON(e.data);
                if (data) {
                    for(var speaker in data) {
                        console.log(data[speaker]);
                        self.stackContent += '<div class="chip">'
                            + data[speaker].name
                        + '</div><br/>';
                    };
                }
                var element = document.getElementById('current-stack');
                element.scrollTop = element.scrollHeight; // Auto scroll to the bottom
            });
        }
    }
});