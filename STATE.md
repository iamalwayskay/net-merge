# State

## 2026-10-28 — Kay privacy fork

- GitHub fork: `iamalwayskay/net-merge`; upstream remains `net-merge/net-merge`.
- Removed the tray's automatic GitHub release request at startup and every six hours.
- No analytics, telemetry or crash-reporting client was present.
- Retained the TCP reachability probes to Cloudflare and Google because the scheduler needs to know
  whether each physical connection can carry traffic. The probes send no application payload.
- Added a regression test that fails if the command imports the automatic updater again.
- Lowered the declared Go version from unavailable 1.27 to the dependency-compatible 1.26.3; verified
  with Ubuntu Go 1.26.8.
- Full test suite passes with `go test ./...`.
- Installed build `0.3.6-kay.1` at `/usr/local/bin/net-merge` with only
  `cap_net_admin,cap_net_raw`; autostart points to that binary and the tray is running.
- The upstream apt binary remains at `/usr/bin/net-merge` as a fallback, but `/usr/local/bin` wins
  on `PATH` and at sign-in.
- This computer has saved credentials for Plusnet, TalkTalk and the phone hotspot, but currently only
  one Wi-Fi radio is exposed. Simultaneous aggregation still requires the phone as USB tethering and
  TalkTalk through Ethernet or a second Wi-Fi adapter.

### Trap

Do not replace the privacy build with the upstream apt binary during routine updates. Rebase from
`upstream/main`, preserve the no-updater regression test, rebuild, and reinstall the local binary.
