# State

## 2026-10-28 — interactive traffic restored to Plusnet

Live diagnosis found that system-wide merging was assigning almost half of new browser flows to
TalkTalk even though its signal was -84 dBm and a 10 MB transfer measured about 0.84 MB/s, versus
5.43 MB/s on Plusnet. Vodafone's Bluetooth PAN was connected but failed the same live transfer.
This made page loads wait on the slow paths rather than adding useful capacity. Paused system-wide
merging and changed the workstation autostart entry to `net-merge tray -off`, so new interactive
connections use Plusnet by default after future sign-ins. TalkTalk and Vodafone remain connected as
available links and merging can still be started explicitly for a bulk multi-connection transfer.
The separate `wg0` interface is the existing WireGuard link to the kdvps private `10.77.0.0/24`
network; it has no default route and therefore was not carrying ordinary internet browsing.

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
- NetworkManager autoconnect is enabled for Plusnet, TalkTalk and `Kaylum's S23+ Network`; Net Merge
  remains auto-paused until at least two physical connections are actually present.
- The MediaTek MT7922 supports two concurrent managed Wi-Fi interfaces. `wlp5s0` is bound to Plusnet
  and `wlp5s0-talktalk` is bound to TalkTalk; both were source-bound and verified with distinct public
  IP addresses. `/etc/systemd/system/net-merge-secondary-wifi.service` recreates the second interface
  at boot, after which NetworkManager autoconnects TalkTalk.
- The phone exposes Bluetooth NAP/PAN; Android accepts it while Bluetooth tethering is enabled.
- With Bluetooth tethering enabled, a direct BlueZ `Network1.Connect("nap")` succeeds. NetworkManager
  then fails because udev renames `bnep0` to a MAC-derived name during DHCP and BlueZ tears it down.
  `packaging/10-bluetooth-pan.link`, installed under `/etc/systemd/network/`, keeps the kernel name.
- Verified all three paths simultaneously: Plusnet on `wlp5s0`, TalkTalk on `wlp5s0-talktalk`, and
  Vodafone mobile data on `bnep0`, each with its own default gateway and public IP. Net Merge assigned
  live flows to all three.
- The workstation side is automatic across login/reboot: the secondary Wi-Fi service, NetworkManager
  autoconnect profiles and Net Merge autostart are enabled. Android must keep Bluetooth tethering on;
  this workstation has no ADB authorization or KDE Connect pairing with which to change that setting.
- `net-merge-mobile-quota.timer` samples `bnep0` every minute, persists both upload and download
  usage under `~/.local/state/net-merge/mobile-quota.json`, warns at 450 GB and disables the Vodafone
  profile at 500 GB per calendar month. It resets the counter and re-enables autoconnect next month;
  Plusnet and TalkTalk are never capped by this guard.

### Trap

Do not replace the privacy build with the upstream apt binary during routine updates. Rebase from
`upstream/main`, preserve the no-updater regression test, rebuild, and reinstall the local binary.
