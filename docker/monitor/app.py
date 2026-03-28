from flask import Flask, render_template_string
import os

app = Flask(__name__)

def get_env(key, fallback):
    return os.getenv(key, fallback)

def get_version():
    return os.getenv("VERSION", "dev")

MAIN_APP_URL = get_env("DEMO_APP_HOST", "http://localhost:3000")

MONITOR_PAGE = """
<html>
<body>
<h2>Monitor Dashboard</h2>

<p><b>Monitor Service Version:</b> {{ monitor_version }}</p>

<h3>Login (use any user created)</h3>
<input id="user" placeholder="username" value="admin"/>
<input id="pass" type="password" placeholder="password" value="admin"/>
<button onclick="login()">Login</button>

<hr/>

<h2>Main App Status</h2>
<pre id="status"></pre>

<h3>List User/Group</h3>
<button onclick="loadUsers()">Load Users</button>

<pre id="users"></pre>

<h3>Delete User</h3>

<input id="deleteUserInput" placeholder="username to delete"/>
<button onclick="deleteUser()">Delete User</button>

<h3>Update Password</h3>
<input id="updateUser" placeholder="username"/>
<input id="updatePass" type="password" placeholder="new password"/>
<button onclick="updatePassword()">Update Password</button>

<script>
async function login() {
    const res = await fetch("{{ main_app_url }}/api/login", {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify({
            username: user.value,
            password: pass.value
        })
    });

    if (!res.ok) {
        alert("Login failed");
        return;
    }

    const data = await res.json();
    localStorage.setItem("token", data.token);
    alert("Logged in!");
}

function getToken() {
    const t = localStorage.getItem("token");
    console.log("TOKEN:", t);
    return t;
}

async function loadUsers() {
    const token = getToken();

    const res = await fetch("{{ main_app_url }}/api/users", {
        method: "GET",
        headers: { 'Authorization': 'Bearer ' + token }
    });

    if (!res.ok) {
        const text = await res.text();
        users.innerText = "Error: " + text;
        return;
    }

    const data = await res.json();
    users.innerText = JSON.stringify(data, null, 2);
}

async function deleteUser() {
    const token = getToken();
    const username = document.getElementById("deleteUserInput").value;

    const res = await fetch("{{ main_app_url }}/api/users/" + username, {
        method: "DELETE",
        headers: {
            'Authorization': 'Bearer ' + token
        }
    });

    if (!res.ok) {
        const text = await res.text();
        alert("Delete failed: " + text);
        return;
    }

    const data = await res.json();
    alert(JSON.stringify(data));
}

async function updatePassword() {
    const token = getToken();
    const username = document.getElementById("updateUser").value;
    const password = document.getElementById("updatePass").value;

    const res = await fetch("{{ main_app_url }}/api/users/" + username, {
        method: "PUT",
        headers: {
            "Authorization": "Bearer " + token,
            "Content-Type": "application/json"
        },
        body: JSON.stringify({ password })
    });

    if (!res.ok) {
        const text = await res.text();
        alert("Update failed: " + text);
        return;
    }

    const data = await res.json();
    alert(JSON.stringify(data));
}

</script>
</body>
</html>
"""

@app.route("/status")
def status():
    return render_template_string(
        MONITOR_PAGE,
        monitor_version=get_version(),
        main_app_url=MAIN_APP_URL
    )

# if __name__ == "__main__":
#     app.run(host="0.0.0.0", port=8000)
