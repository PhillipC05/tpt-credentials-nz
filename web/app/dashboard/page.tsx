"use client";

import { useState, useEffect } from "react";

interface ProfessionalBody {
  id: string;
  name: string;
  slug: string;
  base_url: string;
}

interface Credential {
  id: string;
  user_id: string;
  professional_body_id: string;
  licence_number: string;
  full_name: string;
  status: string;
  verified_at: string | null;
  expires_at: string | null;
  last_checked_at: string | null;
  created_at: string;
  updated_at: string;
}

export default function DashboardPage() {
  const [token, setToken] = useState<string | null>(null);
  const [credentials, setCredentials] = useState<Credential[]>([]);
  const [bodies, setBodies] = useState<ProfessionalBody[]>([]);
  const [selectedSlug, setSelectedSlug] = useState("");
  const [licenceNumber, setLicenceNumber] = useState("");
  const [fullName, setFullName] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [qrData, setQrData] = useState<{ token: string; verify_url: string } | null>(null);

  useEffect(() => {
    const stored = localStorage.getItem("credentials_token");
    if (stored) {
      setToken(stored);
      fetchCredentials(stored);
      fetchBodies(stored);
    }
  }, []);

  const getAuthHeaders = () => ({
    "Content-Type": "application/json",
    Authorization: `Bearer ${token}`,
  });

  const fetchCredentials = async (t: string) => {
    try {
      const res = await fetch("/api/credentials", {
        headers: { Authorization: `Bearer ${t}` },
      });
      if (res.ok) {
        setCredentials(await res.json());
      }
    } catch {
      console.error("Failed to fetch credentials");
    }
  };

  const fetchBodies = async (t: string) => {
    try {
      const res = await fetch("/api/professional-bodies", {
        headers: { Authorization: `Bearer ${t}` },
      });
      if (res.ok) {
        setBodies(await res.json());
      }
    } catch {
      console.error("Failed to fetch professional bodies");
    }
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setSuccess("");

    if (!selectedSlug || !licenceNumber || !fullName) {
      setError("All fields are required");
      return;
    }

    try {
      const res = await fetch("/api/credentials", {
        method: "POST",
        headers: getAuthHeaders(),
        body: JSON.stringify({
          professional_body_slug: selectedSlug,
          licence_number: licenceNumber,
          full_name: fullName,
        }),
      });

      const data = await res.json();
      if (!res.ok) {
        setError(data.error || "Failed to create credential");
        return;
      }

      setSuccess("Credential created successfully! Verification in progress.");
      setLicenceNumber("");
      setFullName("");
      setSelectedSlug("");
      fetchCredentials(token!);
    } catch {
      setError("Network error");
    }
  };

  const handleRefresh = async (id: string) => {
    try {
      const res = await fetch(`/api/credentials/${id}/refresh`, {
        method: "POST",
        headers: getAuthHeaders(),
      });

      if (res.ok) {
        setSuccess("Credential status refreshed!");
        fetchCredentials(token!);
      } else {
        setError("Failed to refresh credential");
      }
    } catch {
      setError("Network error");
    }
  };

  const handleRevoke = async (id: string) => {
    if (!confirm("Are you sure you want to revoke this credential?")) return;

    try {
      const res = await fetch(`/api/credentials/${id}/revoke`, {
        method: "POST",
        headers: getAuthHeaders(),
      });

      if (res.ok) {
        setSuccess("Credential revoked");
        fetchCredentials(token!);
      } else {
        setError("Failed to revoke credential");
      }
    } catch {
      setError("Network error");
    }
  };

  const handleGenerateQR = async (id: string) => {
    try {
      const res = await fetch(`/api/credentials/${id}/qr`, {
        headers: getAuthHeaders(),
      });
      if (res.ok) {
        const data = await res.json();
        setQrData({ token: data.token, verify_url: data.verify_url });
      } else {
        setError("Failed to generate QR code");
      }
    } catch {
      setError("Network error");
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "active":
        return "#27ae60";
      case "pending":
        return "#f39c12";
      case "revoked":
        return "#e74c3c";
      case "expired":
        return "#95a5a6";
      default:
        return "#666";
    }
  };

  if (!token) {
    return (
      <div style={{ textAlign: "center", padding: "4rem 1rem" }}>
        <h2>Please sign in</h2>
        <a href="/" style={{ color: "#3498db" }}>Go to login</a>
      </div>
    );
  }

  return (
    <div>
      <h1 style={{ marginBottom: "0.5rem" }}>Your Credentials</h1>
      <p style={{ color: "#666", marginBottom: "2rem" }}>
        Link your professional licences and generate verification QR codes.
      </p>

      {error && (
        <div style={{ background: "#fde8e8", color: "#c0392b", padding: "0.75rem", borderRadius: "6px", marginBottom: "1rem" }}>
          {error}
        </div>
      )}

      {success && (
        <div style={{ background: "#e8fde8", color: "#27ae60", padding: "0.75rem", borderRadius: "6px", marginBottom: "1rem" }}>
          {success}
        </div>
      )}

      {/* Create Credential Form */}
      <div style={{ background: "#fff", padding: "1.5rem", borderRadius: "12px", boxShadow: "0 2px 8px rgba(0,0,0,0.06)", marginBottom: "2rem" }}>
        <h2 style={{ marginTop: 0, marginBottom: "1rem", fontSize: "1.1rem" }}>Add New Credential</h2>
        <form onSubmit={handleCreate}>
          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr auto", gap: "0.75rem", alignItems: "end" }}>
            <div>
              <label style={{ display: "block", marginBottom: "0.25rem", fontWeight: 600, fontSize: "0.85rem" }}>
                Professional Body
              </label>
              <select
                value={selectedSlug}
                onChange={(e) => setSelectedSlug(e.target.value)}
                required
                style={{
                  width: "100%",
                  padding: "0.5rem",
                  border: "1px solid #ddd",
                  borderRadius: "6px",
                  fontSize: "0.9rem",
                }}
              >
                <option value="">Select...</option>
                {bodies.map((b) => (
                  <option key={b.id} value={b.slug}>
                    {b.name}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label style={{ display: "block", marginBottom: "0.25rem", fontWeight: 600, fontSize: "0.85rem" }}>
                Licence Number
              </label>
              <input
                type="text"
                value={licenceNumber}
                onChange={(e) => setLicenceNumber(e.target.value)}
                required
                placeholder="e.g., MCNZ-12345"
                style={{
                  width: "100%",
                  padding: "0.5rem",
                  border: "1px solid #ddd",
                  borderRadius: "6px",
                  fontSize: "0.9rem",
                  boxSizing: "border-box",
                }}
              />
            </div>

            <div>
              <label style={{ display: "block", marginBottom: "0.25rem", fontWeight: 600, fontSize: "0.85rem" }}>
                Full Name on Licence
              </label>
              <input
                type="text"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                required
                placeholder="e.g., Dr Jane Smith"
                style={{
                  width: "100%",
                  padding: "0.5rem",
                  border: "1px solid #ddd",
                  borderRadius: "6px",
                  fontSize: "0.9rem",
                  boxSizing: "border-box",
                }}
              />
            </div>

            <button
              type="submit"
              style={{
                padding: "0.5rem 1.25rem",
                background: "#1a1a2e",
                color: "#fff",
                border: "none",
                borderRadius: "6px",
                cursor: "pointer",
                fontWeight: 600,
                whiteSpace: "nowrap",
              }}
            >
              Add Credential
            </button>
          </div>
        </form>
      </div>

      {/* Credentials List */}
      {credentials.length === 0 ? (
        <div style={{ textAlign: "center", padding: "3rem", color: "#666" }}>
          <p>No credentials yet. Add your first professional licence above.</p>
        </div>
      ) : (
        <div style={{ display: "grid", gap: "1rem" }}>
          {credentials.map((cred) => (
            <div
              key={cred.id}
              style={{
                background: "#fff",
                padding: "1.25rem",
                borderRadius: "12px",
                boxShadow: "0 2px 8px rgba(0,0,0,0.06)",
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
              }}
            >
              <div>
                <div style={{ fontWeight: 600, marginBottom: "0.25rem" }}>{cred.full_name}</div>
                <div style={{ fontSize: "0.85rem", color: "#666" }}>
                  Licence: {cred.licence_number}
                </div>
                <div style={{ marginTop: "0.25rem" }}>
                  <span
                    style={{
                      display: "inline-block",
                      padding: "0.15rem 0.5rem",
                      borderRadius: "12px",
                      fontSize: "0.75rem",
                      fontWeight: 600,
                      background: getStatusColor(cred.status) + "20",
                      color: getStatusColor(cred.status),
                    }}
                  >
                    {cred.status.toUpperCase()}
                  </span>
                </div>
              </div>
              <div style={{ display: "flex", gap: "0.5rem" }}>
                <button
                  onClick={() => handleRefresh(cred.id)}
                  style={{
                    padding: "0.35rem 0.75rem",
                    background: "#3498db",
                    color: "#fff",
                    border: "none",
                    borderRadius: "6px",
                    cursor: "pointer",
                    fontSize: "0.8rem",
                  }}
                >
                  Refresh
                </button>
                <button
                  onClick={() => handleGenerateQR(cred.id)}
                  style={{
                    padding: "0.35rem 0.75rem",
                    background: "#2ecc71",
                    color: "#fff",
                    border: "none",
                    borderRadius: "6px",
                    cursor: "pointer",
                    fontSize: "0.8rem",
                  }}
                >
                  QR Code
                </button>
                <button
                  onClick={() => handleRevoke(cred.id)}
                  style={{
                    padding: "0.35rem 0.75rem",
                    background: "#e74c3c",
                    color: "#fff",
                    border: "none",
                    borderRadius: "6px",
                    cursor: "pointer",
                    fontSize: "0.8rem",
                  }}
                >
                  Revoke
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* QR Code Modal */}
      {qrData && (
        <div
          onClick={() => setQrData(null)}
          style={{
            position: "fixed",
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            background: "rgba(0,0,0,0.5)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            zIndex: 1000,
          }}
        >
          <div
            onClick={(e) => e.stopPropagation()}
            style={{
              background: "#fff",
              padding: "2rem",
              borderRadius: "12px",
              textAlign: "center",
              maxWidth: "400px",
            }}
          >
            <h3 style={{ marginTop: 0 }}>Verification QR Code</h3>
            <p style={{ fontSize: "0.85rem", color: "#666", marginBottom: "1rem" }}>
              Scan this code to verify the credential. Valid for 30 minutes.
            </p>
            <div
              style={{
                width: "256px",
                height: "256px",
                background: "#f0f0f0",
                margin: "0 auto 1rem",
                borderRadius: "8px",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                color: "#999",
                fontSize: "0.85rem",
              }}
            >
              QR Code Placeholder
              <br />
              (Generated server-side)
            </div>
            <p style={{ fontSize: "0.8rem", color: "#999" }}>
              Verification URL:{" "}
              <a href={qrData.verify_url} style={{ color: "#3498db" }}>
                {qrData.verify_url}
              </a>
            </p>
            <button
              onClick={() => setQrData(null)}
              style={{
                padding: "0.5rem 1.5rem",
                background: "#1a1a2e",
                color: "#fff",
                border: "none",
                borderRadius: "6px",
                cursor: "pointer",
                marginTop: "0.5rem",
              }}
            >
              Close
            </button>
          </div>
        </div>
      )}
    </div>
  );
}