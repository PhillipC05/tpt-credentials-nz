"use client";

import { useState, useEffect } from "react";
import QRVerifyDisplay from "../../components/QRVerifyDisplay";

interface ProfessionalBody {
  id: string;
  name: string;
  slug: string;
}

interface Credential {
  id: string;
  professional_body_id: string;
  licence_number: string;
  full_name: string;
  status: string;
  verified_at: string | null;
  expires_at: string | null;
  last_checked_at: string | null;
  created_at: string;
}

interface CredentialEvent {
  id: string;
  event_type: string;
  from_status: string;
  to_status: string;
  notes: string;
  created_at: string;
}

interface CredentialVisibility {
  credential_id: string;
  show_licence_number: boolean;
  show_expiry: boolean;
  show_verified_at: boolean;
}

interface QRData {
  token: string;
  verify_url: string;
  qr_code_base64: string;
}

const STATUS_COLORS: Record<string, string> = {
  active: "#27ae60",
  pending: "#f39c12",
  revoked: "#e74c3c",
  expired: "#95a5a6",
};

const EVENT_LABELS: Record<string, string> = {
  created: "Credential created",
  verified: "Verified by professional body",
  refreshed: "Status refreshed",
  revoked: "Credential revoked",
  expired: "Credential expired",
  visibility_changed: "Visibility settings updated",
};

export default function DashboardPage() {
  const [token, setToken] = useState<string | null>(null);
  const [credentials, setCredentials] = useState<Credential[]>([]);
  const [bodies, setBodies] = useState<ProfessionalBody[]>([]);
  const [selectedSlug, setSelectedSlug] = useState("");
  const [licenceNumber, setLicenceNumber] = useState("");
  const [fullName, setFullName] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [qrData, setQrData] = useState<QRData | null>(null);
  const [activeTab, setActiveTab] = useState<Record<string, "info" | "timeline" | "privacy">>(
    {}
  );
  const [timelines, setTimelines] = useState<Record<string, CredentialEvent[]>>({});
  const [visibility, setVisibility] = useState<Record<string, CredentialVisibility>>({});

  useEffect(() => {
    const stored = localStorage.getItem("credentials_token");
    if (stored) {
      setToken(stored);
      fetchCredentials(stored);
    }
    fetchBodies();
  }, []);

  const getAuthHeaders = (t?: string) => ({
    "Content-Type": "application/json",
    Authorization: `Bearer ${t ?? token}`,
  });

  const fetchCredentials = async (t: string) => {
    const res = await fetch("/api/credentials", {
      headers: { Authorization: `Bearer ${t}` },
    });
    if (res.ok) setCredentials(await res.json());
  };

  const fetchBodies = async () => {
    const res = await fetch("/api/professional-bodies");
    if (res.ok) setBodies(await res.json());
  };

  const fetchTimeline = async (credId: string) => {
    if (timelines[credId]) return;
    const res = await fetch(`/api/credentials/${credId}/events`, {
      headers: getAuthHeaders()!,
    });
    if (res.ok) {
      const data = await res.json();
      setTimelines((prev) => ({ ...prev, [credId]: data }));
    }
  };

  const fetchVisibility = async (credId: string) => {
    if (visibility[credId]) return;
    const res = await fetch(`/api/credentials/${credId}/visibility`, {
      headers: getAuthHeaders()!,
    });
    if (res.ok) {
      const data = await res.json();
      setVisibility((prev) => ({ ...prev, [credId]: data }));
    }
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setSuccess("");

    const res = await fetch("/api/credentials", {
      method: "POST",
      headers: getAuthHeaders()!,
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

    setSuccess("Credential submitted — verification running in the background.");
    setLicenceNumber("");
    setFullName("");
    setSelectedSlug("");
    fetchCredentials(token!);
  };

  const handleRefresh = async (id: string) => {
    const res = await fetch(`/api/credentials/${id}/refresh`, {
      method: "POST",
      headers: getAuthHeaders()!,
    });
    if (res.ok) {
      setSuccess("Status refreshed");
      fetchCredentials(token!);
      setTimelines((prev) => ({ ...prev, [id]: [] }));
    } else {
      setError("Failed to refresh credential");
    }
  };

  const handleRevoke = async (id: string) => {
    if (!confirm("Revoke this credential? This cannot be undone.")) return;
    const res = await fetch(`/api/credentials/${id}/revoke`, {
      method: "POST",
      headers: getAuthHeaders()!,
    });
    if (res.ok) {
      setSuccess("Credential revoked");
      fetchCredentials(token!);
    } else {
      setError("Failed to revoke credential");
    }
  };

  const handleGenerateQR = async (id: string) => {
    const res = await fetch(`/api/credentials/${id}/qr`, {
      headers: getAuthHeaders()!,
    });
    if (res.ok) {
      setQrData(await res.json());
    } else {
      setError("Failed to generate QR code");
    }
  };

  const handleDownloadVC = async (id: string) => {
    const res = await fetch(`/api/credentials/${id}/vc`, {
      headers: getAuthHeaders()!,
    });
    if (!res.ok) { setError("Failed to export credential"); return; }
    const blob = await res.blob();
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `credential-${id.substring(0, 8)}.jsonld`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const handleVisibilityChange = async (
    credId: string,
    field: keyof CredentialVisibility,
    value: boolean
  ) => {
    const current = visibility[credId] ?? {
      credential_id: credId,
      show_licence_number: true,
      show_expiry: true,
      show_verified_at: true,
    };
    const updated = { ...current, [field]: value };
    setVisibility((prev) => ({ ...prev, [credId]: updated }));

    await fetch(`/api/credentials/${credId}/visibility`, {
      method: "PUT",
      headers: getAuthHeaders()!,
      body: JSON.stringify(updated),
    });
  };

  const setTab = (credId: string, tab: "info" | "timeline" | "privacy") => {
    setActiveTab((prev) => ({ ...prev, [credId]: tab }));
    if (tab === "timeline") fetchTimeline(credId);
    if (tab === "privacy") fetchVisibility(credId);
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
        Link your professional licences and share verified credentials via QR code or W3C Verifiable Credential export.
      </p>

      {error && (
        <div style={{ background: "#fde8e8", color: "#c0392b", padding: "0.75rem", borderRadius: "6px", marginBottom: "1rem" }}>
          {error} <button onClick={() => setError("")} style={{ float: "right", background: "none", border: "none", cursor: "pointer", color: "#c0392b" }}>✕</button>
        </div>
      )}
      {success && (
        <div style={{ background: "#e8fde8", color: "#27ae60", padding: "0.75rem", borderRadius: "6px", marginBottom: "1rem" }}>
          {success} <button onClick={() => setSuccess("")} style={{ float: "right", background: "none", border: "none", cursor: "pointer", color: "#27ae60" }}>✕</button>
        </div>
      )}

      {/* Add Credential Form */}
      <div style={{ background: "#fff", padding: "1.5rem", borderRadius: "12px", boxShadow: "0 2px 8px rgba(0,0,0,0.06)", marginBottom: "2rem" }}>
        <h2 style={{ marginTop: 0, marginBottom: "1rem", fontSize: "1.1rem" }}>Add New Credential</h2>
        <form onSubmit={handleCreate}>
          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr auto", gap: "0.75rem", alignItems: "end" }}>
            <div>
              <label style={{ display: "block", marginBottom: "0.25rem", fontWeight: 600, fontSize: "0.85rem" }}>Professional Body</label>
              <select
                value={selectedSlug}
                onChange={(e) => setSelectedSlug(e.target.value)}
                required
                style={{ width: "100%", padding: "0.5rem", border: "1px solid #ddd", borderRadius: "6px", fontSize: "0.9rem" }}
              >
                <option value="">Select…</option>
                {bodies.map((b) => (
                  <option key={b.id} value={b.slug}>{b.name}</option>
                ))}
              </select>
            </div>
            <div>
              <label style={{ display: "block", marginBottom: "0.25rem", fontWeight: 600, fontSize: "0.85rem" }}>Licence Number</label>
              <input
                type="text"
                value={licenceNumber}
                onChange={(e) => setLicenceNumber(e.target.value)}
                required
                placeholder="e.g., MCNZ-12345"
                style={{ width: "100%", padding: "0.5rem", border: "1px solid #ddd", borderRadius: "6px", fontSize: "0.9rem", boxSizing: "border-box" }}
              />
            </div>
            <div>
              <label style={{ display: "block", marginBottom: "0.25rem", fontWeight: 600, fontSize: "0.85rem" }}>Full Name on Licence</label>
              <input
                type="text"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                required
                placeholder="e.g., Dr Jane Smith"
                style={{ width: "100%", padding: "0.5rem", border: "1px solid #ddd", borderRadius: "6px", fontSize: "0.9rem", boxSizing: "border-box" }}
              />
            </div>
            <button
              type="submit"
              style={{ padding: "0.5rem 1.25rem", background: "#1a1a2e", color: "#fff", border: "none", borderRadius: "6px", cursor: "pointer", fontWeight: 600, whiteSpace: "nowrap" }}
            >
              Add
            </button>
          </div>
        </form>
      </div>

      {/* Credentials */}
      {credentials.length === 0 ? (
        <div style={{ textAlign: "center", padding: "3rem", color: "#666" }}>
          <p>No credentials yet. Add your first professional licence above.</p>
        </div>
      ) : (
        <div style={{ display: "grid", gap: "1rem" }}>
          {credentials.map((cred) => {
            const tab = activeTab[cred.id] ?? "info";
            const statusColor = STATUS_COLORS[cred.status] ?? "#666";
            return (
              <div key={cred.id} style={{ background: "#fff", borderRadius: "12px", boxShadow: "0 2px 8px rgba(0,0,0,0.06)", overflow: "hidden" }}>
                {/* Card header */}
                <div style={{ padding: "1.25rem", display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
                  <div>
                    <div style={{ fontWeight: 600, marginBottom: "0.25rem" }}>{cred.full_name}</div>
                    <div style={{ fontSize: "0.85rem", color: "#666" }}>Licence: {cred.licence_number}</div>
                    {cred.expires_at && (
                      <div style={{ fontSize: "0.8rem", color: "#999", marginTop: "0.2rem" }}>
                        Expires: {new Date(cred.expires_at).toLocaleDateString()}
                      </div>
                    )}
                    <div style={{ marginTop: "0.4rem" }}>
                      <span style={{ display: "inline-block", padding: "0.15rem 0.5rem", borderRadius: "12px", fontSize: "0.75rem", fontWeight: 600, background: statusColor + "20", color: statusColor }}>
                        {cred.status.toUpperCase()}
                      </span>
                    </div>
                  </div>
                  <div style={{ display: "flex", gap: "0.4rem", flexWrap: "wrap", justifyContent: "flex-end" }}>
                    <button onClick={() => handleRefresh(cred.id)} style={btnStyle("#3498db")}>Refresh</button>
                    {cred.status === "active" && (
                      <>
                        <button onClick={() => handleGenerateQR(cred.id)} style={btnStyle("#2ecc71")}>QR Code</button>
                        <button onClick={() => handleDownloadVC(cred.id)} style={btnStyle("#9b59b6")}>Export VC</button>
                      </>
                    )}
                    {cred.status !== "revoked" && (
                      <button onClick={() => handleRevoke(cred.id)} style={btnStyle("#e74c3c")}>Revoke</button>
                    )}
                  </div>
                </div>

                {/* Tabs */}
                <div style={{ borderTop: "1px solid #f0f0f0", display: "flex", fontSize: "0.8rem" }}>
                  {(["info", "timeline", "privacy"] as const).map((t) => (
                    <button
                      key={t}
                      onClick={() => setTab(cred.id, t)}
                      style={{
                        padding: "0.5rem 1rem",
                        border: "none",
                        background: tab === t ? "#f8f9fa" : "transparent",
                        borderBottom: tab === t ? "2px solid #1a1a2e" : "2px solid transparent",
                        cursor: "pointer",
                        fontWeight: tab === t ? 600 : 400,
                        color: tab === t ? "#1a1a2e" : "#666",
                      }}
                    >
                      {t === "info" ? "Details" : t === "timeline" ? "Timeline" : "Privacy"}
                    </button>
                  ))}
                </div>

                {/* Tab content */}
                <div style={{ padding: "1rem 1.25rem", background: "#f8f9fa", fontSize: "0.85rem" }}>
                  {tab === "info" && (
                    <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "0.5rem" }}>
                      <div>
                        <span style={{ color: "#999" }}>Verified at</span>
                        <div>{cred.verified_at ? new Date(cred.verified_at).toLocaleString() : "—"}</div>
                      </div>
                      <div>
                        <span style={{ color: "#999" }}>Last checked</span>
                        <div>{cred.last_checked_at ? new Date(cred.last_checked_at).toLocaleString() : "—"}</div>
                      </div>
                      <div>
                        <span style={{ color: "#999" }}>Created</span>
                        <div>{new Date(cred.created_at).toLocaleString()}</div>
                      </div>
                    </div>
                  )}

                  {tab === "timeline" && (
                    <div>
                      {!timelines[cred.id] ? (
                        <p style={{ color: "#999" }}>Loading…</p>
                      ) : timelines[cred.id].length === 0 ? (
                        <p style={{ color: "#999" }}>No events recorded yet.</p>
                      ) : (
                        <div style={{ display: "flex", flexDirection: "column", gap: "0.5rem" }}>
                          {timelines[cred.id].map((ev) => (
                            <div key={ev.id} style={{ display: "flex", gap: "0.75rem", alignItems: "flex-start" }}>
                              <div style={{ width: "8px", height: "8px", borderRadius: "50%", background: "#1a1a2e", marginTop: "5px", flexShrink: 0 }} />
                              <div>
                                <div style={{ fontWeight: 500 }}>{EVENT_LABELS[ev.event_type] ?? ev.event_type}</div>
                                {ev.notes && <div style={{ color: "#666", fontSize: "0.8rem" }}>{ev.notes}</div>}
                                <div style={{ color: "#999", fontSize: "0.75rem" }}>{new Date(ev.created_at).toLocaleString()}</div>
                              </div>
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  )}

                  {tab === "privacy" && (
                    <div>
                      <p style={{ color: "#666", marginTop: 0, marginBottom: "0.75rem" }}>
                        Control what third parties can see when scanning your QR code or viewing your public profile.
                      </p>
                      {visibility[cred.id] ? (
                        <div style={{ display: "flex", flexDirection: "column", gap: "0.5rem" }}>
                          {(
                            [
                              { key: "show_licence_number", label: "Show licence number" },
                              { key: "show_expiry", label: "Show expiry date" },
                              { key: "show_verified_at", label: "Show verification date" },
                            ] as Array<{ key: keyof CredentialVisibility; label: string }>
                          ).map(({ key, label }) => (
                            <label key={key} style={{ display: "flex", alignItems: "center", gap: "0.5rem", cursor: "pointer" }}>
                              <input
                                type="checkbox"
                                checked={!!visibility[cred.id][key]}
                                onChange={(e) => handleVisibilityChange(cred.id, key, e.target.checked)}
                              />
                              {label}
                            </label>
                          ))}
                        </div>
                      ) : (
                        <p style={{ color: "#999" }}>Loading…</p>
                      )}
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}

      {qrData && (
        <QRVerifyDisplay
          token={qrData.token}
          verifyUrl={qrData.verify_url}
          qrCodeBase64={qrData.qr_code_base64}
          onClose={() => setQrData(null)}
        />
      )}
    </div>
  );
}

function btnStyle(bg: string): React.CSSProperties {
  return {
    padding: "0.35rem 0.75rem",
    background: bg,
    color: "#fff",
    border: "none",
    borderRadius: "6px",
    cursor: "pointer",
    fontSize: "0.8rem",
    fontWeight: 500,
  };
}
