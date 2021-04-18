new Vue({
    el: '#app',

    data: {
        ws: null, // Our websocket
        tableId: null, // meeting ID
        stackContent: '', // A running list of chat messages displayed on the screen
        name: null, // Our username
        joined: false // True if email and username have been filled in
    },

    created: function() {
        var self = this;
        this.joined = false;
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
        send: function () {
            if (this.newMsg != '') {
                this.ws.send(
                    JSON.stringify({
                        tableId: this.tableId,
                        action: this.action,
                        name: this.name
                    }
                ));
            }
        },

        join: function () {
            if (!this.name) {
                Materialize.toast('You must set a meeting ID', 2000);
                return
            }
            this.name = $('<p>').html(this.name).text();
            this.joined = true;
        },

        create: async function () {
            var self = this;
            const requestOptions = {
                method: "POST",
                headers: { "Accept": "application/json" }
            };
            await fetch("http://" + window.location.host + "/ws", requestOptions)
              .then(response => response.json())
              .then(data => self.tableId = data.meetingId);
            this.ws = new WebSocket('ws://' + window.location.host + '/ws?meeting_id=' + this.tableId);
            this.ws.addEventListener('message', function(e) {
                var msg = JSON.parse(e.data);
                self.stackContent += '<div class="chip">'
                        + emojione.toImage(msg.name)
                    + '</div><br/>';

                var element = document.getElementById('current-stack');
                element.scrollTop = element.scrollHeight; // Auto scroll to the bottom
            });
            this.joined = true;
        }
    }
});