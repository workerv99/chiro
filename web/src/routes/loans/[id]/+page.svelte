<script>
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { i18n } from '$lib/i18n.svelte.js';
  import { api } from '$lib/api.svelte.js';
  import { payInstallment, cascadeInstallment, unpayInstallment, remove, updateLoan } from '$lib/stores.svelte.js';
  import { money, toDisplay, toISO, todayISO } from '$lib/format.js';
  import ConfirmSheet from '$lib/components/ConfirmSheet.svelte';

  const loanId = String(page.params.id);
  let loan = $state(null);
  let schedule = $state([]);
  let loading = $state(true);
  let payAmount = $state('');
  let payDate = $state(toDisplay(todayISO()));
  let confirmDel = $state(false);
  let confirmPay = $state(false);
  let pendingAction = $state(null);

  let showEdit = $state(false);
  let editForm = $state({
    description: '',
    amount: '',
    date: '',
    interest_rate: '0',
    interest_type: 'simple',
    months: '1',
    frequency: 'monthly',
    first_due_date: ''
  });
  let editErr = $state('');
  let editDay = $state('');
  let editMonth = $state('');
  let editYear = $state('');
  let editDueDay = $state('');
  let editDueMonth = $state('');
  let editDueYear = $state('');

  async function load() {
    const [loans, s] = await Promise.all([api('/api/loans'), api(`/api/loans/${loanId}/installments`)]);
    loan = loans.find((l) => l.loan_id === loanId) || null;
    schedule = s;
    if (nextPending) payAmount = String(nextPending.amount);
  }

  function openEdit() {
    if (!loan) return;
    const d = new Date(loan.date);
    editDay = String(d.getDate()).padStart(2, '0');
    editMonth = String(d.getMonth() + 1).padStart(2, '0');
    editYear = String(d.getFullYear());

    if (loan.first_due_date) {
      const fd = new Date(loan.first_due_date);
      editDueDay = String(fd.getDate()).padStart(2, '0');
      editDueMonth = String(fd.getMonth() + 1).padStart(2, '0');
      editDueYear = String(fd.getFullYear());
    } else {
      editDueDay = '';
      editDueMonth = '';
      editDueYear = '';
    }

    editForm = {
      description: loan.description || '',
      amount: String(loan.amount),
      date: loan.date,
      interest_rate: String(loan.interest_rate || 0),
      interest_type: loan.interest_type || 'simple',
      months: String(loan.months || 1),
      frequency: loan.frequency || 'monthly',
      first_due_date: loan.first_due_date || ''
    };
    editErr = '';
    showEdit = true;
  }

  async function saveEdit() {
    editErr = '';
    const amt = parseFloat(editForm.amount);
    if (!amt || amt <= 0) return (editErr = 'Monto invalido');
    if (!editDay || !editMonth || !editYear) return (editErr = 'Fecha requerida');

    const dateStr = `${editYear}-${editMonth}-${editDay}`;
    let firstDueStr = null;
    if (editDueDay && editDueMonth && editDueYear) {
      firstDueStr = `${editDueYear}-${editDueMonth}-${editDueDay}`;
    }

    try {
      await updateLoan(loanId, {
        description: editForm.description.trim(),
        amount: amt,
        date: dateStr,
        interest_rate: parseFloat(editForm.interest_rate) || 0,
        interest_type: editForm.interest_type,
        months: parseInt(editForm.months, 10) || 1,
        frequency: editForm.frequency,
        first_due_date: firstDueStr
      });
      showEdit = false;
      await load();
    } catch (e) {
      editErr = e.message;
    }
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

  async function generatePDF() {
    const [{ jsPDF }, autoTable] = await Promise.all([
      import('jspdf'),
      import('jspdf-autotable')
    ]);
    const doc = new jsPDF();
    const pageWidth = doc.internal.pageSize.getWidth();
    const pageHeight = doc.internal.pageSize.getHeight();

    const brand = { primary: [91, 124, 246], dark: [30, 41, 59], light: [241, 245, 249] };
    const colors = { green: [22, 163, 74], red: [220, 38, 38], amber: [245, 158, 11], gray: [100, 116, 139] };
    const margin = 14;
    const contentWidth = pageWidth - margin * 2;

    const drawHeader = () => {
      doc.setFillColor(...brand.dark);
      doc.rect(0, 0, pageWidth, 32, 'F');
      doc.setFillColor(...brand.primary);
      doc.rect(0, 32, pageWidth, 2, 'F');

      doc.setTextColor(255, 255, 255);
      doc.setFontSize(8);
      doc.setFont('helvetica', 'normal');
      doc.text('CHIRO', margin, 12);
      doc.setFontSize(18);
      doc.setFont('helvetica', 'bold');
      doc.text('REPORTE DE PRESTAMO', margin, 24);
    };

    const drawFooter = (pageNum) => {
      doc.setFillColor(...brand.light);
      doc.rect(0, pageHeight - 16, pageWidth, 16, 'F');
      doc.setFontSize(7);
      doc.setFont('helvetica', 'normal');
      doc.setTextColor(...colors.gray);
      doc.text(`Generado por Chiro · ${new Date().toLocaleString('es-EC')}`, margin, pageHeight - 6);
      doc.text(`Pagina ${pageNum}`, pageWidth - margin, pageHeight - 6, { align: 'right' });
    };

    drawHeader();
    drawFooter(1);

    let y = 44;

    doc.setTextColor(...brand.dark);
    doc.setFontSize(11);
    doc.setFont('helvetica', 'bold');
    doc.text(loan.person_name, margin, y);
    y += 6;

    doc.setFontSize(9);
    doc.setFont('helvetica', 'normal');
    doc.setTextColor(...colors.gray);
    const freqLabel = { monthly: 'Mensual', biweekly: 'Quincenal', weekly: 'Semanal' };
    doc.text(`${loan.description || 'Sin descripcion'} · ${freqLabel[loan.frequency] || loan.frequency} · ${toDisplay(loan.date)}`, margin, y);
    y += 10;

    doc.setDrawColor(...brand.light);
    doc.setLineWidth(0.5);
    doc.line(margin, y, pageWidth - margin, y);
    y += 8;

    const interestAmount = (loan.total_amount * (loan.interest_rate || 0)) / 100;
    const totalWithInterest = loan.total_amount + interestAmount;
    const progressPct = loan.total_amount > 0 ? (loan.total_paid / loan.total_amount * 100).toFixed(0) : 0;

    const summaryCards = [
      { label: 'CAPITAL', value: `$${loan.total_amount.toFixed(2)}`, color: brand.dark },
      { label: 'INTERES', value: `$${interestAmount.toFixed(2)}`, sub: `${loan.interest_rate || 0}%`, color: colors.gray },
      { label: 'TOTAL', value: `$${totalWithInterest.toFixed(2)}`, color: brand.primary },
      { label: 'SALDO', value: `$${remaining.toFixed(2)}`, color: remaining > 0 ? colors.red : colors.green }
    ];

    const cardWidth = (contentWidth - 18) / 4;
    summaryCards.forEach((card, i) => {
      const x = margin + i * (cardWidth + 6);
      doc.setFillColor(250, 250, 250);
      doc.roundedRect(x, y, cardWidth, 24, 2, 2, 'F');
      doc.setDrawColor(...brand.light);
      doc.roundedRect(x, y, cardWidth, 24, 2, 2, 'S');

      doc.setFontSize(7);
      doc.setFont('helvetica', 'bold');
      doc.setTextColor(...colors.gray);
      doc.text(card.label, x + 5, y + 7);

      doc.setFontSize(14);
      doc.setFont('helvetica', 'bold');
      doc.setTextColor(...card.color);
      doc.text(card.value, x + 5, y + 17);

      if (card.sub) {
        doc.setFontSize(7);
        doc.setFont('helvetica', 'normal');
        doc.setTextColor(...colors.gray);
        doc.text(card.sub, x + 5, y + 22);
      }
    });
    y += 32;

    doc.setFillColor(245, 245, 245);
    doc.roundedRect(margin, y, contentWidth, 12, 2, 2, 'F');
    doc.setFontSize(8);
    doc.setFont('helvetica', 'normal');
    doc.setTextColor(...brand.dark);
    doc.text(`${schedule.length} cuotas · Abonado: $${loan.total_paid.toFixed(2)} · Progreso: ${progressPct}%`, margin + 5, y + 7.5);
    y += 18;

    const barY = y;
    const barHeight = 6;
    doc.setFillColor(230, 230, 230);
    doc.roundedRect(margin, barY, contentWidth, barHeight, 3, 3, 'F');
    if (loan.total_amount > 0) {
      const fillWidth = Math.max(0, Math.min(contentWidth, (loan.total_paid / loan.total_amount) * contentWidth));
      if (fillWidth > 0) {
        doc.setFillColor(...colors.green);
        doc.roundedRect(margin, barY, fillWidth, barHeight, 3, 3, 'F');
      }
    }
    y += barHeight + 12;

    const tableData = schedule.map((s, idx) => {
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
        if (diff > 0) delay = `${diff} dia(s)`;
        else delay = 'A tiempo';
      } else if (!isPaid) {
        const due = new Date(s.due_date);
        const today = new Date();
        const diff = Math.ceil((today - due) / (1000 * 60 * 60 * 24));
        if (diff > 0) delay = `${diff} dia(s)`;
        else delay = 'Pendiente';
      }

      return [s.number, dueDate, amount, paidDate, paidAmt, status, delay];
    });

    autoTable(doc, {
      startY: y,
      head: [['#', 'VENCIMIENTO', 'MONTO', 'PAGO', 'ABONADO', 'ESTADO', 'DIAS']],
      body: tableData,
      theme: 'striped',
      headStyles: {
        fillColor: brand.primary,
        textColor: 255,
        fontStyle: 'bold',
        fontSize: 7,
        cellPadding: 3
      },
      styles: {
        fontSize: 8,
        cellPadding: 3.5,
        textColor: brand.dark,
        lineColor: [230, 230, 230],
        lineWidth: 0.3
      },
      alternateRowStyles: {
        fillColor: [248, 250, 252]
      },
      columnStyles: {
        0: { cellWidth: 10, halign: 'center', fontStyle: 'bold' },
        1: { cellWidth: 24 },
        2: { cellWidth: 20, halign: 'right' },
        3: { cellWidth: 22 },
        4: { cellWidth: 20, halign: 'right' },
        5: { cellWidth: 18, halign: 'center', fontStyle: 'bold' },
        6: { cellWidth: 18, halign: 'center' }
      },
      didParseCell: (data) => {
        if (data.section === 'body') {
          const row = data.row;
          if (data.column.index === 5) {
            const val = row.raw[5];
            if (val === 'Saldada') {
              data.cell.styles.textColor = colors.green;
              data.cell.styles.fontStyle = 'bold';
            } else {
              data.cell.styles.textColor = colors.gray;
            }
          }
          if (data.column.index === 6) {
            const val = row.raw[6];
            if (val.includes('dia(s)')) {
              data.cell.styles.textColor = colors.red;
            } else if (val === 'A tiempo') {
              data.cell.styles.textColor = colors.green;
            }
          }
        }
      }
    });

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
    <div class="inline-flex">
      <button class="btn btn-small" onclick={openEdit}>Editar</button>
      <button class="btn btn-small" onclick={generatePDF}>PDF</button>
    </div>
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
    {#if nextPending}
      <div style="margin-bottom:12px">
        <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:4px">
          <span style="font-size:0.82rem;color:var(--ink-dim);font-weight:600">Cuota #{nextPending.number}</span>
          <span style="font-size:0.82rem;color:var(--ink-dim)">{toDisplay(nextPending.due_date)}</span>
        </div>
        <div style="font-size:1.3rem;font-weight:800">{money(nextPending.amount)}</div>
      </div>
      <button class="btn btn-primary" style="width:100%;margin-bottom:12px" onclick={() => pay(true)}>
        Pagar cuota #{nextPending.number}
      </button>
    {:else}
      <div style="text-align:center;padding:12px 0;margin-bottom:12px">
        <div style="color:var(--green);font-weight:700;margin-bottom:4px">Todas las cuotas pagadas</div>
        <div style="font-size:0.82rem;color:var(--ink-dim)">${loan.total_paid.toFixed(2)} de ${loan.total_amount.toFixed(2)}</div>
      </div>
    {/if}

    <details class="more-options">
      <summary>Pago personalizado</summary>
      <div style="margin-top:12px">
        <div class="grid2" style="margin-bottom:10px">
          <div class="form-field">
            <label class="eyebrow" for="pay-amount">Importe</label>
            <input id="pay-amount" bind:value={payAmount} inputmode="decimal" />
          </div>
          <div class="form-field">
            <label class="eyebrow" for="pay-date">Fecha</label>
            <input id="pay-date" bind:value={payDate} placeholder="DD/MM/YYYY" />
          </div>
        </div>
        <button class="btn" style="width:100%" onclick={() => pay(false)} disabled={!nextPending || !payAmount}>
          Aplicar pago personalizado
        </button>
      </div>
    </details>

    {#if paidCount > 0}
      <div style="margin-top:12px;padding-top:12px;border-top:1px solid var(--border)">
        <button class="btn btn-cancel" style="width:100%;font-size:0.82rem" onclick={unpay}>
          Deshacer ultimo pago
        </button>
      </div>
    {/if}
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

{#if showEdit}
  <div class="overlay" onclick={() => (showEdit = false)}>
    <div class="sheet" role="dialog" aria-modal="true" aria-labelledby="edit-loan-title" onclick={(e) => e.stopPropagation()}>
      <div class="handle"></div>
      <h2 class="title" id="edit-loan-title" style="margin-bottom:16px">Editar prestamo</h2>

      <div class="form-field">
        <label class="eyebrow" for="edit-desc">Descripcion *</label>
        <input id="edit-desc" bind:value={editForm.description} />
      </div>

      <div class="form-field">
        <label class="eyebrow" for="edit-amount">Monto *</label>
        <input id="edit-amount" bind:value={editForm.amount} inputmode="decimal" />
      </div>

      <div class="form-field">
        <label class="eyebrow">Fecha *</label>
        <div class="grid3">
          <div>
            <label class="eyebrow" style="font-size:0.7rem" for="edit-day">Dia</label>
            <input id="edit-day" bind:value={editDay} placeholder="DD" maxlength="2" inputmode="numeric" />
          </div>
          <div>
            <label class="eyebrow" style="font-size:0.7rem" for="edit-month">Mes</label>
            <input id="edit-month" bind:value={editMonth} placeholder="MM" maxlength="2" inputmode="numeric" />
          </div>
          <div>
            <label class="eyebrow" style="font-size:0.7rem" for="edit-year">Ano</label>
            <input id="edit-year" bind:value={editYear} placeholder="AAAA" maxlength="4" inputmode="numeric" />
          </div>
        </div>
      </div>

      <div class="form-field">
        <label class="eyebrow" for="edit-rate">Tasa de interes (%)</label>
        <input id="edit-rate" bind:value={editForm.interest_rate} inputmode="decimal" placeholder="0" />
        <p class="hint-text">Ej: $1000 x 10% = $100 de interes</p>
      </div>

      <div class="form-field">
        <label class="eyebrow" for="edit-months">Numero de cuotas</label>
        <input id="edit-months" bind:value={editForm.months} inputmode="numeric" />
      </div>

      <div class="form-field">
        <label class="eyebrow">Frecuencia</label>
        <div class="freq-tabs">
          <button class="freq-tab" class:active={editForm.frequency === 'monthly'} onclick={() => editForm.frequency = 'monthly'}>Mensual</button>
          <button class="freq-tab" class:active={editForm.frequency === 'biweekly'} onclick={() => editForm.frequency = 'biweekly'}>Quincenal</button>
          <button class="freq-tab" class:active={editForm.frequency === 'weekly'} onclick={() => editForm.frequency = 'weekly'}>Semanal</button>
        </div>
      </div>

      <div class="form-field">
        <label class="eyebrow">Primera cuota</label>
        <div class="grid3">
          <div>
            <label class="eyebrow" style="font-size:0.7rem" for="edit-due-day">Dia</label>
            <input id="edit-due-day" bind:value={editDueDay} placeholder="DD" maxlength="2" inputmode="numeric" />
          </div>
          <div>
            <label class="eyebrow" style="font-size:0.7rem" for="edit-due-month">Mes</label>
            <input id="edit-due-month" bind:value={editDueMonth} placeholder="MM" maxlength="2" inputmode="numeric" />
          </div>
          <div>
            <label class="eyebrow" style="font-size:0.7rem" for="edit-due-year">Ano</label>
            <input id="edit-due-year" bind:value={editDueYear} placeholder="AAAA" maxlength="4" inputmode="numeric" />
          </div>
        </div>
        <p class="hint-text">{editForm.months || 1} cuotas x ${money(parseFloat(editForm.amount) || 0)}</p>
      </div>

      {#if editErr}
        <p class="error-text" role="alert">{editErr}</p>
      {/if}

      <div class="inline-flex" style="margin-top:8px">
        <button class="btn btn-cancel" onclick={() => (showEdit = false)}>Cancelar</button>
        <button class="btn btn-primary" onclick={saveEdit}>Guardar</button>
      </div>
    </div>
  </div>
{/if}
