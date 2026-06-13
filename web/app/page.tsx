"use client";

import { useState, useEffect } from "react";

interface ProfessionalBody {
  id: string;
  name: string;
  slug: string;
}

export default function HomePage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [name, setName] = useState("");
  const [isRegistering, setIsRegistering] = useState(false);
  const [token, setToken] = useState<string | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    const stored = localStorage.getItem("credentials_token");
    if (stored) setToken(stored);
  }, []);

  const handleAuth = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    const endpoint = isRegistering ? "/api/auth/register" : "/api/auth/login";
    const body = isRegistering
      ? { email, password, name }
      : { email, password };

    try {
      const res = await fetch(endpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });

      const data = await res.json();
      if (!res.ok) {
        setError(data.error || "Authentication failed");
        return;
      }

      localStorage.setItem("credentials_token", data.token);
      setToken(data.token);
    } catch {
      setError("Network error");
    }
  };

  const handleLogout = () => {
    localStorage.removeItem("credentials_token");
    setToken(null);
  };

  if (token) {
    return (
      <div style={{ textAlign: "center", padding: "4rem 1rem" }}>
        <h1 style={{ fontSize: "2rem", marginBottom: "1rem" }}>Professional Credentials Wallet</h1>
        <p style={{ color: "#666", marginBottom: "2rem", fontSize: "1.1rem" }}>
          Manage and share verified professional credentials linked to your RealMe identity.
        </p>
        <div style={{ display: "flex", gap: "1rem", justifyContent: "center" }}>
          <a
            href="/dashboard"
            style={{
              padding: "0.75rem 2rem",
              background: "#1a1a2e",
              color: "#fff",
              textDecoration: "none",
              borderRadius: "8px",
              fontWeight: 600,
            }}
          >
            Go to Dashboard
          </a>
          <button
            onClick={handleLogout}
            style={{
              padding: "0.75rem 2rem",
              background: "#e74c3c",
              color: "#fff",
              border: "none",
              borderRadius: "8px",
              cursor: "pointer",
              fontWeight: 600,
            }}
          >
            Logout
          </button>
        </div>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: "420px", margin: "4rem auto", padding: "2rem", background: "#fff", borderRadius: "12px", boxShadow: "0 2px 12px rgba(0,0,0,0.08)" }}>
      <h1 style={{ textAlign: "center", marginBottom: "0.5rem" }}>Credentials Wallet</h1>
      <p style={{ textAlign: "center", color: "#666", marginBottom: "2rem", fontSize: "0.9rem" }}>
        Sign in to manage your professional credentials
      </p>

      <form onSubmit={handleAuth}>
        <div style={{ marginBottom: "1rem" }}>
          <label style={{ display: "block", marginBottom: "0.25rem", fontWeight: 600 }}>
            Email
          </label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            style={{
              width: "100%",
              padding: "0.6rem",
              border: "1px solid #ddd",
              borderRadius: "6px",
              fontSize: "1rem",
              boxSizing: "border-box",
            }}
          />
        </div>

        {isRegistering && (
          <div style={{ marginBottom: "1rem" }}>
            <label style={{ display: "block", marginBottom: "0.25rem", fontWeight: 600 }}>
              Full Name
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              style={{
                width: "100%",
                padding: "0.6rem",
                border: "1px solid #ddd",
                borderRadius: "6px",
                fontSize: "1rem",
                boxSizing: "border-box",
              }}
            />
          </div>
        )}

        <div style={{ marginBottom: "1rem" }}>
          <label style={{ display: "block", marginBottom: "0.25rem", fontWeight: 600 }}>
            Password
          </label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            style={{
              width: "100%",
              padding: "0.6rem",
              border: "1px solid #ddd",
              borderRadius: "6px",
              fontSize: "1rem",
              boxSizing: "border-box",
            }}
          />
        </div>

        {error && (
          <p style={{ color: "#e74c3c", fontSize: "0.875rem", marginBottom: "1rem" }}>{error}</p>
        )}

        <button
          type="submit"
          style={{
            width: "100%",
            padding: "0.75rem",
            background: "#1a1a2e",
            color: "#fff",
            border: "none",
            borderRadius: "6px",
            fontSize: "1rem",
            cursor: "pointer",
            fontWeight: 600,
          }}
        >
          {isRegistering ? "Register" : "Sign In"}
        </button>
      </form>

      <p style={{ textAlign: "center", marginTop: "1.5rem", color: "#666", fontSize: "0.9rem" }}>
        {isRegistering ? "Already have an account?" : "Don't have an account?"}{" "}
        <button
          onClick={() => setIsRegistering(!isRegistering)}
          style={{ background: "none", border: "none", color: "#3498db", cursor: "pointer", textDecoration: "underline" }}
        >
          {isRegistering ? "Sign In" : "Register"}
        </button>
      </p>
    </div>
  );
}