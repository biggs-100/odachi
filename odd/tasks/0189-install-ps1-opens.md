# install-ps1-opens
Base: eff512dfce0a8cb2b190b80df1329aebb6800a6d
Light: yes

## Requirements (verbatim, from the user)
R1 · "install.ps1 opens with #Requires, not a byte-order mark: the documented irm | iex install prints no error"

## Done when

## Steps
- [x] T1 · install.ps1 opens with #Requires, not a byte-order mark: the documented irm | iex install prints no error   R1 · tree:5222b4b765326a2c7c61d2baa35437ee3866cc47

## History
L1 · "install.ps1 opens with #Requires, not a byte-order mark: the documented irm | iex install prints no error"
L2 · OFICIO · "install.ps1 is served without a BOM: the documented irm | iex path decodes UTF-8 and prints the messages right, and PowerShell 5.1 reading the file from disk pays with ANSI-decoded accents" (porque the command the README teaches is the one a user runs, and a first line that errors teaches them to ignore errors; the direct-file path is undocumented and still runs)
