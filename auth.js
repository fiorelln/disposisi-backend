// public/js/auth.js

async function login() {
    const email = document.getElementById("email").value;
    const pass = document.getElementById("password").value;

    const res = await apiPost("login", {
        email: email,
        password: pass
    });

    if (res.success) {
        alert("Login berhasil!");
        localStorage.setItem("token", res.data.token);
        window.location.href = "dashboard.html";
    } else {
        alert(res.data.error || "Login gagal!");
    }
}


async function registerUser() {
    const name = document.getElementById("name").value;
    const email = document.getElementById("email").value;
    const pass = document.getElementById("password").value;

    const role = document.getElementById("role").value; // kalau ada dropdown

    const res = await apiPost("register", {
        name: name,
        email: email,
        password: pass,
        role: role
    });

    if (res.success) {
        alert("Registrasi berhasil!");
        window.location.href = "login.html";
    } else {
        alert(res.data.error || "Gagal daftar!");
    }
}
