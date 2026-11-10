import React, { useState } from 'react';
import { PaymentTransaction } from '../types';
import { DollarSign, CreditCard, RefreshCw, AlertCircle, CheckCircle, ShieldCheck, Download, ArrowUpRight } from 'lucide-react';
import { Language } from '../utils/translations';

interface FinancePortalProps {
  transactions: PaymentTransaction[];
  dashboard: any;
  onRefresh: () => void;
  lang: Language;
}

export const FinancePortal: React.FC<FinancePortalProps> = ({
  transactions,
  dashboard,
  onRefresh,
  lang,
}) => {
  const [selectedTx, setSelectedTx] = useState<PaymentTransaction | null>(null);
  const [refundAmount, setRefundAmount] = useState<number>(0);
  const [refundReason, setRefundReason] = useState<string>('Operational schedule conflict / patient cancellation');
  const [isRefunding, setIsRefunding] = useState<boolean>(false);

  const metrics = dashboard?.metrics || {
    total_revenue: 0,
    paid_count: 0,
    bkash_revenue: 0,
    ssl_revenue: 0,
    refunded_count: 0,
  };

  const handleProcessRefund = async () => {
    if (!selectedTx || refundAmount <= 0) return;
    setIsRefunding(true);
    try {
      const token = localStorage.getItem('upasham_token') || '';
      const res = await fetch('/api/finance/refund', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          payment_id: selectedTx.id,
          amount: Number(refundAmount),
          reason: refundReason,
        }),
      });

      if (res.ok) {
        setSelectedTx(null);
        onRefresh();
      }
    } catch (e) {
      console.error(e);
    } finally {
      setIsRefunding(false);
    }
  };

  return (
    <div style={{ padding: '36px 0', backgroundColor: '#F7FBFC', minHeight: '80vh' }}>
      <div className="container">
        {/* Header */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '28px', flexWrap: 'wrap', gap: '16px' }}>
          <div>
            <span style={{ fontSize: '12px', fontWeight: 800, color: '#33C2DF', textTransform: 'uppercase' }}>
              FINANCE & GATEWAY RECONCILIATION
            </span>
            <h1 style={{ fontSize: '28px', fontWeight: 800, color: '#2A3F68' }}>
              Platform Transactions & Settlements
            </h1>
          </div>

          <div style={{ display: 'flex', gap: '10px' }}>
            <button onClick={onRefresh} className="btn btn-outline" style={{ fontSize: '13px' }}>
              <RefreshCw size={14} /> Refresh Ledger
            </button>
          </div>
        </div>

        {/* Financial Stat Cards */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: '18px', marginBottom: '32px' }}>
          <div className="card-safeer" style={{ padding: '22px' }}>
            <span style={{ fontSize: '11px', color: '#64748B', fontWeight: 700, textTransform: 'uppercase' }}>
              Total Gross Revenue
            </span>
            <div style={{ fontSize: '28px', fontWeight: 800, color: '#2A3F68', marginTop: '4px' }}>
              ৳{metrics.total_revenue.toLocaleString()} <span style={{ fontSize: '12px', color: '#64748B' }}>BDT</span>
            </div>
            <span style={{ fontSize: '12px', color: '#33C2DF', fontWeight: 600 }}>
              {metrics.paid_count} verified paid orders
            </span>
          </div>

          <div className="card-safeer" style={{ padding: '22px' }}>
            <span style={{ fontSize: '11px', color: '#64748B', fontWeight: 700, textTransform: 'uppercase' }}>
              Direct bKash Volume
            </span>
            <div style={{ fontSize: '28px', fontWeight: 800, color: '#E2136E', marginTop: '4px' }}>
              ৳{metrics.bkash_revenue.toLocaleString()} <span style={{ fontSize: '12px', color: '#64748B' }}>BDT</span>
            </div>
            <span style={{ fontSize: '12px', color: '#64748B' }}>Tokenized gateway adapter</span>
          </div>

          <div className="card-safeer" style={{ padding: '22px' }}>
            <span style={{ fontSize: '11px', color: '#64748B', fontWeight: 700, textTransform: 'uppercase' }}>
              SSLCOMMERZ Volume
            </span>
            <div style={{ fontSize: '28px', fontWeight: 800, color: '#9B468A', marginTop: '4px' }}>
              ৳{metrics.ssl_revenue.toLocaleString()} <span style={{ fontSize: '12px', color: '#64748B' }}>BDT</span>
            </div>
            <span style={{ fontSize: '12px', color: '#64748B' }}>Cards & MFS gateway adapter</span>
          </div>

          <div className="card-safeer" style={{ padding: '22px' }}>
            <span style={{ fontSize: '11px', color: '#64748B', fontWeight: 700, textTransform: 'uppercase' }}>
              Processed Refunds
            </span>
            <div style={{ fontSize: '28px', fontWeight: 800, color: '#D97706', marginTop: '4px' }}>
              {metrics.refunded_count}
            </div>
            <span style={{ fontSize: '12px', color: '#64748B' }}>Controlled ledger adjustment</span>
          </div>
        </div>

        {/* Transactions Table */}
        <div style={{ marginBottom: '32px' }}>
          <h3 style={{ fontSize: '18px', fontWeight: 800, color: '#2A3F68', marginBottom: '14px' }}>
            Gateway Transactions Ledger ({transactions.length})
          </h3>

          <div className="card-safeer" style={{ overflow: 'hidden', padding: 0 }}>
            <div style={{ overflowX: 'auto' }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '13.5px' }}>
                <thead style={{ backgroundColor: '#EEF7F8', borderBottom: '1px solid #DCEAEC', color: '#475569' }}>
                  <tr>
                    <th style={{ padding: '14px 18px' }}>Invoice Ref</th>
                    <th style={{ padding: '14px 18px' }}>Patient & Booking</th>
                    <th style={{ padding: '14px 18px' }}>Gateway</th>
                    <th style={{ padding: '14px 18px' }}>Amount</th>
                    <th style={{ padding: '14px 18px' }}>Trx ID</th>
                    <th style={{ padding: '14px 18px' }}>Status</th>
                    <th style={{ padding: '14px 18px' }}>Date</th>
                    <th style={{ padding: '14px 18px', textAlign: 'right' }}>Action</th>
                  </tr>
                </thead>
                <tbody>
                  {transactions.map((tx) => (
                    <tr key={tx.id} style={{ borderBottom: '1px solid #EEF7F8' }}>
                      <td style={{ padding: '14px 18px', fontWeight: 700, color: '#2A3F68' }}>
                        {tx.invoice_ref}
                      </td>

                      <td style={{ padding: '14px 18px' }}>
                        <strong>{tx.patient_name}</strong>
                        <div style={{ fontSize: '11px', color: '#64748B' }}>
                          Ref: {tx.booking_ref} ({tx.service_type})
                        </div>
                      </td>

                      <td style={{ padding: '14px 18px' }}>
                        <span
                          style={{
                            fontWeight: 700,
                            color: tx.gateway === 'bkash' ? '#E2136E' : '#9B468A',
                          }}
                        >
                          {tx.gateway?.toUpperCase()}
                        </span>
                      </td>

                      <td style={{ padding: '14px 18px', fontWeight: 800, color: '#2A3F68' }}>
                        ৳{tx.amount} BDT
                      </td>

                      <td style={{ padding: '14px 18px', fontFamily: 'monospace', fontSize: '12px' }}>
                        {tx.gateway_tx_id || 'PENDING'}
                      </td>

                      <td style={{ padding: '14px 18px' }}>
                        <span
                          className={`badge ${
                            tx.status === 'Paid'
                              ? 'badge-green'
                              : tx.status === 'Refunded'
                              ? 'badge-red'
                              : 'badge-yellow'
                          }`}
                        >
                          {tx.status}
                        </span>
                      </td>

                      <td style={{ padding: '14px 18px', fontSize: '12px', color: '#64748B' }}>
                        {new Date(tx.created_at).toLocaleString()}
                      </td>

                      <td style={{ padding: '14px 18px', textAlign: 'right' }}>
                        {tx.status === 'Paid' && (
                          <button
                            onClick={() => {
                              setSelectedTx(tx);
                              setRefundAmount(tx.amount);
                            }}
                            className="btn btn-outline"
                            style={{ padding: '5px 10px', fontSize: '12px', color: '#DC2626' }}
                          >
                            Refund
                          </button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>

        {/* Modal: Process Refund */}
        {selectedTx && (
          <div
            style={{
              position: 'fixed',
              inset: 0,
              backgroundColor: 'rgba(42,63,104,0.62)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              zIndex: 300,
              padding: '20px',
            }}
          >
            <div className="card-safeer animate-fade-in" style={{ width: '100%', maxWidth: '440px', padding: '28px', backgroundColor: '#FFF' }}>
              <h3 style={{ fontSize: '18px', fontWeight: 800, color: '#DC2626', marginBottom: '6px' }}>
                Issue Authorized Refund
              </h3>
              <p style={{ fontSize: '13px', color: '#64748B', marginBottom: '18px' }}>
                Invoice: {selectedTx.invoice_ref} • Patient: {selectedTx.patient_name}
              </p>

              <div style={{ marginBottom: '14px' }}>
                <label style={{ fontSize: '12px', fontWeight: 700, display: 'block', marginBottom: '4px' }}>
                  Refund Amount (BDT):
                </label>
                <input
                  type="number"
                  max={selectedTx.amount}
                  value={refundAmount}
                  onChange={(e) => setRefundAmount(Number(e.target.value))}
                  style={{ width: '100%', padding: '10px 12px', borderRadius: '8px', border: '1px solid #C7D8DE' }}
                />
              </div>

              <div style={{ marginBottom: '20px' }}>
                <label style={{ fontSize: '12px', fontWeight: 700, display: 'block', marginBottom: '4px' }}>
                  Reason for Refund (Audit Logged):
                </label>
                <input
                  type="text"
                  value={refundReason}
                  onChange={(e) => setRefundReason(e.target.value)}
                  style={{ width: '100%', padding: '10px 12px', borderRadius: '8px', border: '1px solid #C7D8DE' }}
                />
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '10px' }}>
                <button onClick={() => setSelectedTx(null)} className="btn btn-outline">
                  Cancel
                </button>
                <button onClick={handleProcessRefund} disabled={isRefunding} className="btn btn-primary" style={{ backgroundColor: '#DC2626', borderColor: '#DC2626' }}>
                  {isRefunding ? 'Processing...' : 'Approve Refund'}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

