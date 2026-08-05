<script>
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { i18n } from '$lib/i18n.svelte.js';
  import { api } from '$lib/api.svelte.js';
  import { payInstallment, cascadeInstallment, unpayInstallment, remove, updateLoan } from '$lib/stores.svelte.js';
  import { money, toDisplay, toISO, todayISO } from '$lib/format.js';
  import ConfirmSheet from '$lib/components/ConfirmSheet.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Card from '$lib/components/ui/card.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Dialog from '$lib/components/ui/dialog.svelte';
  import DialogHeader from '$lib/components/ui/dialog-header.svelte';
  import DialogTitle from '$lib/components/ui/dialog-title.svelte';
  import DialogFooter from '$lib/components/ui/dialog-footer.svelte';
  import { ChevronLeft, Edit, FileText } from 'lucide-svelte';

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
    description: '', amount: '', date: '', interest_rate: '0',
    interest_type: 'simple', months: '1', frequency: 'monthly',
    first_due_date: '', custom_installment: '0'
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
    } else { editDueDay = ''; editDueMonth = ''; editDueYear = ''; }
    editForm = {
      description: loan.description || '', amount: String(loan.amount),
      date: loan.date, interest_rate: String(loan.interest_rate || 0),
      interest_type: loan.interest_type || 'simple', months: String(loan.months || 1),
      frequency: loan.frequency || 'monthly', first_due_date: loan.first_due_date || '',
      custom_installment: '0'
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
    if (editDueDay && editDueMonth && editDueYear) firstDueStr = `${editDueYear}-${editDueMonth}-${editDueDay}`;
    try {
      await updateLoan(loanId, {
        description: editForm.description.trim(), amount: amt, date: dateStr,
        interest_rate: parseFloat(editForm.interest_rate) || 0,
        interest_type: editForm.interest_type, months: parseInt(editForm.months, 10) || 1,
        frequency: editForm.frequency, first_due_date: firstDueStr,
        custom_installment: parseFloat(editForm.custom_installment) || 0
      });
      showEdit = false;
      await load();
    } catch (e) { editErr = e.message; }
  }

  $effect(() => { loading = true; load().finally(() => (loading = false)); });

  const nextPending = $derived(schedule.find((x) => !x.is_paid));
  const paidCount = $derived(schedule.reduce((c, x) => c + (x.is_paid ? 1 : 0), 0));
  const remaining = $derived(loan ? loan.total_amount - loan.total_paid : 0);
  const progressPct = $derived(loan && loan.total_amount > 0 ? Math.round((loan.total_paid / loan.total_amount) * 100) : 0);

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

  async function generatePDF() {
    const [{ jsPDF }, autoTableModule] = await Promise.all([import('jspdf'), import('jspdf-autotable')]);
    const autoTable = autoTableModule.default || autoTableModule;
    const doc = new jsPDF();
    const pw = doc.internal.pageSize.getWidth();
    const m = 14;
    const cw = pw - m * 2;

    doc.setFillColor(30, 41, 59);
    doc.rect(0, 0, pw, 32, 'F');
    doc.setFillColor(91, 124, 246);
    doc.rect(0, 32, pw, 2, 'F');

    doc.setTextColor(255, 255, 255);
    doc.setFontSize(7);
    doc.setFont('helvetica', 'normal');
    doc.text('CHIRO', m, 12);
    doc.setFontSize(18);
    doc.setFont('helvetica', 'bold');
    doc.text('REPORTE DE PRESTAMO', m, 24);

    let y = 44;
    doc.setTextColor(30, 41, 59);
    doc.setFontSize(11);
    doc.setFont('helvetica', 'bold');
    doc.text(loan.person_name, m, y);
    y += 6;

    doc.setFontSize(9);
    doc.setFont('helvetica', 'normal');
    doc.setTextColor(100, 116, 139);
    doc.text(`${loan.description || 'Sin descripcion'} · ${loan.frequency}`, m, y);
    y += 5;
    doc.text(`Ejecutado: ${toDisplay(loan.date)} · Primera cuota: ${toDisplay(loan.first_due_date)}`, m, y);
    y += 10;

    doc.setDrawColor(230, 230, 230);
    doc.setLineWidth(0.5);
    doc.line(m, y, pw - m, y);
    y += 8;

    const interestAmount = (loan.total_amount * (loan.interest_rate || 0)) / 100;
    const cards = [
      { label: 'CAPITAL', value: `$${loan.total_amount.toFixed(2)}` },
      { label: 'INTERES', value: `$${interestAmount.toFixed(2)}`, sub: `${loan.interest_rate || 0}%` },
      { label: 'TOTAL', value: `$${(loan.total_amount + interestAmount).toFixed(2)}` },
      { label: 'SALDO', value: `$${remaining.toFixed(2)}`, color: remaining > 0 ? [239, 68, 68] : [34, 197, 94] }
    ];

    const cardW = (cw - 18) / 4;
    cards.forEach((c, i) => {
      const x = m + i * (cardW + 6);
      doc.setFillColor(250, 250, 250);
      doc.roundedRect(x, y, cardW, 24, 2, 2, 'F');
      doc.setDrawColor(230, 230, 230);
      doc.roundedRect(x, y, cardW, 24, 2, 2, 'S');
      doc.setFontSize(7);
      doc.setFont('helvetica', 'bold');
      doc.setTextColor(100, 116, 139);
      doc.text(c.label, x + 5, y + 7);
      doc.setFontSize(14);
      doc.setFont('helvetica', 'bold');
      doc.setTextColor(c.color ? c.color[0] : 0, c.color ? c.color[1] : 0, c.color ? c.color[2] : 0);
      doc.text(c.value, x + 5, y + 17);
    });
    y += 32;

    const progressPct = loan.total_amount > 0 ? Math.round((loan.total_paid / loan.total_amount) * 100) : 0;
    doc.setFillColor(245, 245, 245);
    doc.roundedRect(m, y, cw, 12, 2, 2, 'F');
    doc.setFontSize(8);
    doc.setFont('helvetica', 'normal');
    doc.setTextColor(30, 41, 59);
    doc.text(`${schedule.length} cuotas · Abonado: $${loan.total_paid.toFixed(2)} · Progreso: ${progressPct}%`, m + 5, y + 7.5);
    y += 18;

    doc.setFillColor(230, 230, 230);
    doc.roundedRect(m, y, cw, 6, 3, 3, 'F');
    if (loan.total_amount > 0) {
      const fillW = Math.max(0, Math.min(cw, (loan.total_paid / loan.total_amount) * cw));
      if (fillW > 0) {
        doc.setFillColor(34, 197, 94);
        doc.roundedRect(m, y, fillW, 6, 3, 3, 'F');
      }
    }
    y += 18;

    const tableData = schedule.map((s) => {
      const dueDate = toDisplay(s.due_date);
      const paidDate = s.paid_date ? toDisplay(s.paid_date) : '—';
      const amount = `$${s.amount.toFixed(2)}`;
      const paidAmt = s.is_paid ? `$${(s.paid_amount || s.amount).toFixed(2)}` : '$0.00';
      const status = s.is_paid ? 'Saldada' : 'Impaga';
      let delay = '—';
      if (s.is_paid && s.paid_date && s.due_date) {
        const diff = Math.ceil((new Date(s.paid_date) - new Date(s.due_date)) / 864e5);
        delay = diff > 0 ? `${diff} dia(s)` : 'A tiempo';
      } else if (!s.is_paid) {
        const diff = Math.ceil((new Date() - new Date(s.due_date)) / 864e5);
        delay = diff > 0 ? `${diff} dia(s)` : 'Pendiente';
      }
      return [s.number, dueDate, amount, paidDate, paidAmt, status, delay];
    });

    autoTable(doc, {
      startY: y,
      head: [['#', 'VENCIMIENTO', 'MONTO', 'PAGO', 'ABONADO', 'ESTADO', 'DIAS']],
      body: tableData,
      theme: 'striped',
      tableWidth: 'auto',
      headStyles: { fillColor: accent, textColor: 255, fontStyle: 'bold', fontSize: 7 },
      styles: { fontSize: 8, cellPadding: 3.5, textColor: [30, 41, 59], lineColor: [230, 230, 230], lineWidth: 0.3 },
      alternateRowStyles: { fillColor: [248, 250, 252] },
      columnStyles: {
        0: { halign: 'center', fontStyle: 'bold' },
        2: { halign: 'right' },
        4: { halign: 'right' },
        5: { halign: 'center', fontStyle: 'bold' },
        6: { halign: 'center' }
      },
      didParseCell: (data) => {
        if (data.section === 'body') {
          const row = data.row;
          if (data.column.index === 5) {
            const val = row.raw[5];
            if (val === 'Saldada') data.cell.styles.textColor = [34, 197, 94];
            else data.cell.styles.textColor = [100, 116, 139];
          }
          if (data.column.index === 6) {
            const val = row.raw[6];
            if (val.includes('dia(s)')) data.cell.styles.textColor = [239, 68, 68];
            else if (val === 'A tiempo') data.cell.styles.textColor = [34, 197, 94];
          }
        }
      }
    });

    doc.save(`prestamo-${loan.person_name.replace(/\s+/g, '_')}-${loanId}.pdf`);
  }
</script>

<svelte:head><title>{loan ? loan.person_name : 'Préstamo'} · Chiro</title></svelte:head>

{#if loading || !loan}
  <p class="text-sm text-muted-foreground py-8 text-center">{i18n.t('common.loading')}</p>
{:else}
  <a class="text-sm text-muted-foreground hover:text-foreground font-semibold mb-2 inline-block" href={`/loans/person/${loan.person_id}`}>← {loan.person_name}</a>
  <h1 class="text-2xl font-bold mb-4">{loan.description || 'Préstamo'}</h1>

  <Card class="p-4 mb-4">
    <div class="grid grid-cols-2 sm:grid-cols-3 gap-4 text-sm">
      <div><span class="text-muted-foreground">Persona</span><p class="font-semibold">{loan.person_name}</p></div>
      <div><span class="text-muted-foreground">Capital</span><p class="font-semibold">{money(loan.amount)}</p></div>
      <div><span class="text-muted-foreground">Ejecutado</span><p class="font-semibold">{toDisplay(loan.date)}</p></div>
      <div><span class="text-muted-foreground">Primera cuota</span><p class="font-semibold">{toDisplay(loan.first_due_date)}</p></div>
      <div><span class="text-muted-foreground">Frecuencia</span><p class="font-semibold">{loan.frequency}</p></div>
      <div><span class="text-muted-foreground">Cuotas</span><p class="font-semibold">{paidCount}/{schedule.length}</p></div>
    </div>
    <div class="flex gap-2 mt-4 pt-4 border-t">
      <Button variant="outline" size="sm" onclick={openEdit}><Edit class="h-4 w-4 mr-1" /> Editar</Button>
      <Button variant="outline" size="sm" onclick={generatePDF}><FileText class="h-4 w-4 mr-1" /> PDF</Button>
    </div>
  </Card>

  <Card class="p-4 mb-4">
    <p class="text-xs font-bold text-muted-foreground uppercase mb-1">Pendiente</p>
    <p class="text-3xl font-extrabold" class:text-destructive={remaining > 0} class:text-green-500={remaining <= 0}>{money(remaining)}</p>
    <p class="text-xs text-muted-foreground mt-1">{money(loan.total_paid)} pagado de {money(loan.total_amount)}</p>
    <div class="h-1.5 bg-border rounded-full overflow-hidden mt-3">
      <div class="h-full bg-green-500 rounded-full transition-all" style="width:{progressPct}%"></div>
    </div>
    <div class="flex justify-between text-xs text-muted-foreground mt-1">
      <span>{paidCount}/{schedule.length} cuotas</span>
      <span>{progressPct}%</span>
    </div>
  </Card>

    {#if nextPending}
      {@const overdueClass = nextPending.is_overdue ? 'border-destructive/35' : ''}
      <Card class="p-4 mb-4 {overdueClass}">
      {#if nextPending.is_overdue}
        <div class="flex items-center gap-2 mb-3">
          <span class="inline-flex items-center rounded-full bg-destructive/10 px-2 py-0.5 text-xs font-bold text-destructive">Vencida</span>
          <span class="text-sm text-destructive font-semibold">
            {Math.ceil((new Date() - new Date(nextPending.due_date)) / (1000 * 60 * 60 * 24))} dia(s) de retraso
          </span>
        </div>
      {/if}
      <div class="flex justify-between items-center mb-2">
        <span class="text-sm text-muted-foreground font-semibold">Cuota #{nextPending.number}</span>
        <span class="text-sm text-muted-foreground">{toDisplay(nextPending.due_date)}</span>
      </div>
      <p class="text-2xl font-extrabold mb-4">{money(nextPending.amount)}</p>
      <Button class="w-full h-12" onclick={() => pay(true)}>Pagar cuota #{nextPending.number}</Button>
      <details class="mt-4">
        <summary class="cursor-pointer text-sm font-semibold text-muted-foreground py-2">Pago personalizado</summary>
        <div class="space-y-3 pt-2">
          <div class="grid grid-cols-2 gap-3">
            <div><Label class="text-xs">Importe</Label><Input bind:value={payAmount} inputmode="decimal" /></div>
            <div><Label class="text-xs">Fecha</Label><Input bind:value={payDate} placeholder="DD/MM/YYYY" /></div>
          </div>
          <Button class="w-full" variant="outline" onclick={() => pay(false)} disabled={!payAmount}>Aplicar pago</Button>
        </div>
      </details>
      {#if paidCount > 0}
        <Button variant="ghost" class="w-full mt-3 text-sm text-muted-foreground" onclick={unpay}>Deshacer ultimo pago</Button>
      {/if}
    </Card>
  {:else}
    <Card class="p-4 mb-4 text-center">
      <p class="font-bold text-green-500 mb-1">Todas las cuotas pagadas</p>
      <p class="text-sm text-muted-foreground">{money(loan.total_paid)} de {money(loan.total_amount)}</p>
    </Card>
  {/if}

  <Card class="overflow-hidden mb-4">
    <h3 class="font-bold p-4 pb-0">Calendario de cuotas</h3>
    {#each schedule as s (s.number)}
      {@const rowClass = !s.is_paid && s.is_overdue ? 'bg-red-500/5' : ''}
      <div class="flex items-center gap-3 px-4 py-3 border-t first:border-t-0 {rowClass}">
          <div class="h-2.5 w-2.5 rounded-full" style="background:{s.is_paid ? '#22c55e' : s.is_overdue ? '#ef4444' : '#94a3b8'}"></div>
        <div class="flex-1">
          <p class="text-sm font-semibold">
            Cuota {s.number}
            {#if !s.is_paid && s.is_overdue}
              <span class="inline-flex items-center rounded-full bg-destructive/10 px-2 py-0.5 text-[10px] font-bold text-destructive ml-1">Vencida</span>
            {/if}
          </p>
          <p class="text-xs text-muted-foreground">
            Vence: {toDisplay(s.due_date)}
            {#if s.is_paid && s.paid_date}<span class="text-green-500 ml-1">Pagado: {toDisplay(s.paid_date)}</span>{/if}
          </p>
        </div>
        <div class="text-right">
          <p class="text-sm font-bold">{money(s.amount)}</p>
          <p class="text-xs">
            {#if s.is_paid}<span class="text-green-500 font-semibold">Saldada</span>
            {:else if s.is_partial}<span class="text-amber-500 font-semibold">Parcial</span>
            {:else if s.is_overdue}<span class="text-destructive font-semibold">Pendiente</span>
            {:else}<span class="text-muted-foreground">Pendiente</span>{/if}
          </p>
        </div>
      </div>
    {/each}
  </Card>

  <Button variant="destructive" class="w-full mb-8" onclick={() => (confirmDel = true)}>Eliminar préstamo</Button>
{/if}

<Dialog bind:open={showEdit}>
  <DialogHeader><DialogTitle>Editar préstamo</DialogTitle></DialogHeader>
  <div class="space-y-4">
    <div class="space-y-2"><Label>Descripción</Label><Input bind:value={editForm.description} /></div>
    <div class="grid grid-cols-2 gap-3">
      <div class="space-y-2"><Label>Monto</Label><Input bind:value={editForm.amount} inputmode="decimal" /></div>
      <div class="space-y-2"><Label>Cuotas</Label><Input bind:value={editForm.months} inputmode="numeric" /></div>
    </div>
    <div class="space-y-2"><Label>Cuota personalizada</Label><Input bind:value={editForm.custom_installment} inputmode="decimal" placeholder="0 = cuota igual" /></div>
    <div class="space-y-2"><Label>Frecuencia</Label>
      <select bind:value={editForm.frequency} class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
        <option value="monthly">Mensual</option><option value="biweekly">Quincenal</option><option value="weekly">Semanal</option>
      </select>
    </div>
    {#if editErr}<p class="text-sm text-destructive">{editErr}</p>{/if}
  </div>
  <DialogFooter>
    <Button variant="outline" onclick={() => (showEdit = false)}>Cancelar</Button>
    <Button onclick={saveEdit}>Guardar</Button>
  </DialogFooter>
</Dialog>

{#if confirmDel}
  <ConfirmSheet bind:open={() => confirmDel, (v) => confirmDel = v} title="Eliminar préstamo" message="¿Estás seguro? Se eliminará el préstamo y su calendario." confirmLabel="Eliminar" danger onConfirm={doDelete} onCancel={() => (confirmDel = false)} />
{/if}

{#if confirmPay && pendingAction}
  <ConfirmSheet bind:open={() => confirmPay, (v) => confirmPay = v} title={pendingAction.type === 'cascade' ? 'Reestructurar' : 'Pagar'} message={`¿Confirmar pago de $${pendingAction.amount.toFixed(2)}?`} confirmLabel="Pagar" onConfirm={executeAction} onCancel={() => { confirmPay = false; pendingAction = null; }} />
{/if}