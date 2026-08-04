<script>
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { i18n } from '$lib/i18n.svelte.js';
  import { api } from '$lib/api.svelte.js';
  import { payInstallment, cascadeInstallment, unpayInstallment, remove } from '$lib/stores.svelte.js';
  import { money, toDisplay, toISO, todayISO } from '$lib/format.js';
  import ConfirmSheet from '$lib/components/ConfirmSheet.svelte';
  import { jsPDF } from 'jspdf';
  import 'jspdf-autotable';

  const loanId = String(page.params.id);
  let loan = $state(null);
  let schedule = $state([]);
  let loading = $state(true);
  let payAmount = $state('');
  let payDate = $state(toDisplay(todayISO()));
  let confirmDel = $state(false);
  let confirmPay = $state(false);
  let pendingAction = $state(null);

  async function load() {
    const [loans, s] = await Promise.all([api('/api/loans'), api(`/api/loans/${loanId}/installments`)]);
    loan = loans.find((l) => l.loan_id === loanId) || null;
    schedule = s;
    if (nextPending) payAmount = String(nextPending.amount);
  }

  $effect(() => {
    loading = true;
    load().finally(() => (loading = false));
  });

  const nextPending = $derived(schedule.find((x) => !x.is_paid));
  const paidCount = $derived(schedule.filter((x) => x.is_paid).length);
  const remaining = $derived(loan ? loan.total_amount - loan.total_paid : 0);
  const blastRadius = $derived(`${schedule.length} ${i18n.t('loans.progress').toLowerCase()} (${paidCount} ${i18n.t('loans.paid')})`);

  async function pay(next) {
    const target = nextPending;
    if (!target) return;
    const amt = next ? target.amount : parseFloat(payAmount);
    if (!amt || amt <= 0) return;
    pendingAction = { type: 'pay', amount: amt, date: toISO(payDate) };
    confirmPay = true;
  }

  async function executePay() {
    if (!pendingAction || pendingAction.type !== 'pay') return;
    const target = nextPending;
    if (!target) return;
    confirmPay = false;
    await payInstallment(target.installment_id, { amount: pendingAction.amount, date: pendingAction.date });
    pendingAction = null;
    await load();
  }

  async function cascade() {
    const target = nextPending;
    if (!target) return;
    const amt = parseFloat(payAmount);
    if (!amt || amt <= 0) return;
    pendingAction = { type: 'cascade', amount: amt, date: toISO(payDate) };
    confirmPay = true;
  }

  async function executeCascade() {
    if (!pendingAction || pendingAction.type !== 'cascade') return;
    const target = nextPending;
    if (!target) return;
    confirmPay = false;
    await cascadeInstallment(target.installment_id, { amount: pendingAction.amount, date: pendingAction.date });
    pendingAction = null;
    await load();
  }

  async function executeAction() {
    if (pendingAction?.type === 'pay') return executePay();
    if (pendingAction?.type === 'cascade') return executeCascade();
  }

  async function unpay() {
    const last = [...schedule].reverse().find((x) => x.is_paid);
    if (!last) return;
    await unpayInstallment(last.installment_id);
    await load();
  }

  async function doDelete() {
    confirmDel = false;
    await remove('loans', loanId);
    goto('/loans');
  }

  function onKeydown(e) {
    if (e.key === 'Escape') {
      if (confirmDel) confirmDel = false;
      else if (confirmPay) { confirmPay = false; pendingAction = null; }
    }
  }

  function generatePDF() {
    const doc = new jsPDF();
    const pageWidth = doc.internal.pageSize.getWidth();

    const accent = [91, 124, 246];
    const green = [34, 197, 94];
    const red = [239, 68, 68];
    const gray = [128, 128, 128];

    doc.setFontSize(22);
    doc.setFont('helvetica', 'bold');
    doc.text('Reporte de prestamo', 14, 20);

    doc.setFontSize(10);
    doc.setFont('helvetica', 'normal');
    doc.setTextColor(gray[0], gray[1], gray[2]);
    const freqLabel = { monthly: 'Mensual', biweekly: 'Quincenal', weekly: 'Semanal' };
    doc.text(`${loan.person_name} · ${loan.description || 'Sin descripcion'} · prestado el ${toDisplay(loan.date)}`, 14, 28);
    doc.setTextColor(0, 0, 0);

    const cardY = 34;
    const cardH = 22;
    const cardW = (pageWidth - 28 - 18) / 4;
    const cards = [
      { label: 'CAPITAL', value: `$${loan.total_amount.toFixed(2)}` },
      { label: `INTERES (${loan.interest_rate || 0}%)`, value: `$${((loan.total_amount * (loan.interest_rate || 0)) / 100).toFixed(2)}` },
      { label: 'TOTAL A PAGAR', value: `$${(loan.total_amount + (loan.total_amount * (loan.interest_rate || 0)) / 100).toFixed(2)}` },
      { label: 'SALDO', value: `$${remaining.toFixed(2)}`, highlight: true }
    ];

    cards.forEach((c, i) => {
      const x = 14 + i * (cardW + 6);
      doc.setFillColor(248, 248, 248);
      doc.roundedRect(x, cardY, cardW, cardH, 3, 3, 'F');
      doc.setFontSize(7);
      doc.setFont('helvetica', 'bold');
      doc.setTextColor(gray[0], gray[1], gray[2]);
      doc.text(c.label, x + 6, cardY + 7);
      doc.setFontSize(13);
      doc.setFont('helvetica', 'bold');
      if (c.highlight) {
        doc.setTextColor(red[0], red[1], red[2]);
      } else {
        doc.setTextColor(0, 0, 0);
      }
      doc.text(c.value, x + 6, cardY + 16);
      doc.setTextColor(0, 0, 0);
    });

    doc.setFontSize(10);
    doc.setFont('helvetica', 'normal');
    doc.setTextColor(0, 0, 0);
    doc.text(`${schedule.length} cuota(s) · frecuencia ${freqLabel[loan.frequency] || loan.frequency} · abonado $${loan.total_paid.toFixed(2)}`, 14, cardY + cardH + 10);

    const tableStartY = cardY + cardH + 18;
    const tableData = schedule.map((s) => {
      const isPaid = s.is_paid;
      const dueDate = toDisplay(s.due_date);
      const paidDate = s.paid_date ? toDisplay(s.paid_date) : '—';
      const amount = `$${s.amount.toFixed(2)}`;
      const paidAmt = isPaid ? `$${(s.paid_amount || s.amount).toFixed(2)}` : '$0.00';
      const status = isPaid ? 'Saldada' : 'Impaga';

      let delay = '—';
      if (isPaid && s.paid_date && s.due_date) {
        const due = new Date(s.due_date);
        const paid = new Date(s.paid_date);
        const diff = Math.ceil((paid - due) / (1000 * 60 * 60 * 24));
        if (diff > 0) delay = `${diff} dia(s) tarde`;
        else delay = 'a tiempo';
      } else if (!isPaid) {
        const due = new Date(s.due_date);
        const today = new Date();
        const diff = Math.ceil((today - due) / (1000 * 60 * 60 * 24));
        if (diff > 0) delay = `${diff} dia(s) tarde`;
        else delay = 'a tiempo';
      }

      return [s.number, dueDate, amount, paidDate, paidAmt, status, delay];
    });

    doc.autoTable({
      startY: tableStartY,
      head: [['#', 'VENCE', 'MONTO', 'PAGO', 'ABONADO', 'ESTADO', 'ATRASO']],
      body: tableData,
      theme: 'grid',
      headStyles: {
        fillColor: accent,
        textColor: 255,
        fontStyle: 'bold',
        fontSize: 8
      },
      styles: {
        fontSize: 8,
        cellPadding: 3,
        textColor: 40
      },
      columnStyles: {
        0: { cellWidth: 10, halign: 'center' },
        1: { cellWidth: 22 },
        2: { cellWidth: 20, halign: 'right' },
        3: { cellWidth: 22 },
        4: { cellWidth: 20, halign: 'right' },
        5: { cellWidth: 18, halign: 'center' },
        6: { cellWidth: 24, halign: 'center' }
      },
      didParseCell: (data) => {
        if (data.section === 'body') {
          const row = data.row;
          const statusIdx = 5;
          const delayIdx = 6;
          if (data.column.index === statusIdx) {
            const val = row.raw[statusIdx];
            if (val === 'Saldada') {
              data.cell.styles.textColor = green;
              data.cell.styles.fontStyle = 'bold';
            } else {
              data.cell.styles.textColor = gray;
            }
          }
          if (data.column.index === delayIdx) {
            const val = row.raw[delayIdx];
            if (val.includes('tarde')) {
              data.cell.styles.textColor = red;
            } else if (val === 'a tiempo') {
              data.cell.styles.textColor = green;
            }
          }
        }
      }
    });

    const finalY = doc.lastAutoTable.finalY + 10;
    doc.setFontSize(8);
    doc.setFont('helvetica', 'normal');
    doc.setTextColor(gray[0], gray[1], gray[2]);
    doc.text(`Generado por Chiro · ${new Date().toLocaleString('es-EC')}`, 14, finalY);

    doc.save(`prestamo-${loan.person_name.replace(/\s+/g, '_')}-${loanId}.pdf`);
  }
</script>

<svelte:head><title>{loan ? loan.person_name : i18n.t('loans.title')} · Chiro</title></svelte:head>
<svelte:window onkeydown={onKeydown} />

{#if loading || !loan}
  <p class="meta" style="padding:24px;text-align:center">{i18n.t('common.loading')}</p>
{:else}
  <div class="page-head">
    <div>
      <a class="back-link" href={`/loans/person/${loan.person_id}`}>← {loan.person_name}</a>
      <h1 class="headline">{loan.description || i18n.t('loans.loan')}</h1>
      <p class="meta">
        {#if loan.frequency}{loan.frequency}{/if}
      </p>
    </div>
    <button class="btn btn-small" onclick={generatePDF}>📄 PDF</button>
  </div>

  <div class="summary-cards">
    <div class="card stat-card">
      <div class="stat-label">{i18n.t('loans.remaining')}</div>
      <div class="stat-value">{money(remaining)}</div>
    </div>
    <div class="card stat-card">
      <div class="stat-label">{i18n.t('loans.paidLabel')}</div>
      <div class="stat-value">{money(loan.total_paid)}</div>
    </div>
    <div class="card stat-card">
      <div class="stat-label">{i18n.t('loans.progress')}</div>
      <div class="stat-value">{paidCount}/{schedule.length}</div>
    </div>
  </div>

  <div class="card" style="padding:16px;margin-top:14px">
    <div class="grid2" style="align-items:end">
      <div class="form-field">
        <label for="pay-amount">{i18n.t('loans.paymentAmount')}</label>
        <input id="pay-amount" bind:value={payAmount} inputmode="decimal" />
      </div>
      <div class="form-field">
        <label for="pay-date">{i18n.t('loans.paymentDate')}</label>
        <input id="pay-date" bind:value={payDate} placeholder={i18n.t('expenses.datePlaceholder')} />
      </div>
    </div>
    <div class="inline-flex" style="margin-top:8px;flex-wrap:wrap">
      <button class="btn btn-primary" onclick={() => pay(true)} disabled={!nextPending}>
        {i18n.t('loans.payNext')}
      </button>
      <button class="btn" onclick={() => pay(false)} disabled={!nextPending}>{i18n.t('loans.payNow')}</button>
      <button class="btn" onclick={cascade} disabled={!nextPending}>{i18n.t('loans.cascade')}</button>
      <button class="btn" onclick={unpay} disabled={paidCount === 0}>{i18n.t('loans.unpay')}</button>
    </div>
  </div>

  <div class="card list-card" style="margin-top:14px">
    <h3 class="card-title" style="padding:12px 16px 0">{i18n.t('loans.schedule')}</h3>
    {#each schedule as s (s.number)}
      <div class="row">
        <div class="cat-dot" style="background:{s.is_paid ? 'var(--green)' : s.is_overdue ? 'var(--red)' : 'var(--ink-dim)'}"></div>
        <div class="row-body">
          <div class="row-title">{i18n.t('loans.installmentN')} {s.number}</div>
          <div class="row-sub">{toDisplay(s.due_date)}</div>
        </div>
        <div class="row-right">
          <div class="row-amount">{money(s.amount)}</div>
          <div class="row-sub">
            {#if s.is_paid}
              <span class="badge-ok">{i18n.t('loans.paid')}</span>
            {:else if s.is_partial}
              <span class="badge-warn">{i18n.t('loans.partial')}</span>
            {:else}
              <span class="badge-pending">{i18n.t('loans.pending')}</span>
            {/if}
          </div>
        </div>
      </div>
    {/each}
  </div>

  <button class="btn btn-danger" style="margin:16px 0 32px" onclick={() => (confirmDel = true)}>{i18n.t('common.delete')}</button>
{/if}

{#if confirmDel}
  <ConfirmSheet
    title={i18n.t('common.delete')}
    message={`${i18n.t('loans.deleteConfirm')}\n\n${blastRadius}`}
    confirmLabel={i18n.t('common.delete')}
    danger
    onConfirm={doDelete}
    onCancel={() => (confirmDel = false)}
  />
{/if}

{#if confirmPay && pendingAction}
  <ConfirmSheet
    title={pendingAction.type === 'cascade' ? i18n.t('loans.cascade') : i18n.t('loans.payNow')}
    message={`¿Confirmar pago de $${pendingAction.amount.toFixed(2)}?`}
    confirmLabel={i18n.t('common.save')}
    onConfirm={executeAction}
    onCancel={() => { confirmPay = false; pendingAction = null; }}
  />
{/if}
