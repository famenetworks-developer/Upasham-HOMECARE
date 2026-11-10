export async function downloadAuthenticatedPDF(id: string) {
  const token = localStorage.getItem('upasham_token');
  if (!token) throw new Error('Please sign in to download this prescription.');
  const response = await fetch(`/api/prescriptions/${encodeURIComponent(id)}/pdf`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!response.ok) throw new Error('Prescription download was not authorized.');
  const blob = await response.blob();
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = `Upasham_Prescription_${id}.pdf`;
  document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();
  URL.revokeObjectURL(url);
}
