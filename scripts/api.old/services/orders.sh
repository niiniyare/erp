#!/usr/bin/env bash
# Orders service tests — stub showing how to add a new module.
#
# To implement:
#   1. Add your endpoints under run_orders()
#   2. run.sh auto-discovers this file — no registration needed

MENU+=("orders:Orders — [stub, add your tests here]:run_orders")

run_orders() {
  header "Orders — (module not yet implemented)"
  warn "This is a stub. Implement tests in scripts/api/services/orders.sh"
  info "See services/users.sh for a working example."
  info "Pattern:"
  info "  resp=\$(apiv GET /api/v1/orders)"
  info "  split_resp \"\$resp\""
  info "  assert_any \"GET /api/v1/orders\" \"\$RESP_STATUS\" \"200\" \"403\""
}
