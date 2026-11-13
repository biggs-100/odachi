# el-ratchet-mira
Base: 5f2db71772508e1fb3d8e8953f22566d7a89bb7c

## Requirements (verbatim, from the user)
R1 · "el ratchet mira la negativa que el humano lee, no el constructor que la arma: la de review que hoy no dice que correr pasa a decirlo, y el mecanismo la vigila"

## Done when
- una negativa que un helper arma sin llamar a blocked entra al ratchet: o esta en el baseline o el suite falla
- el walker lee los constructores que el paquete declara, y un paquete que declara uno conserva su conteo: check y start no se mueven
- la frase de refused de review dice que correr, como ya lo dice la de check
- el baseline sigue bajando solo: una frase pagada se borra y una nueva se agrega a proposito
- go test ./... is green and odachi check --run "go test ./..." is green
- ## Unreleased names it, patch (0292)

## Steps
- [x] T1 · el walker recibe los constructores que el paquete declara, y cada paquete que hoy lee blocked sigue igual   R1 · route:inline · tree:c5d308c3fe1ff4731c9beb968fc72e530c08d87f
- [x] T2 · review declara refused y su baseline lo congela, sin mover el conteo de check ni de start   R1 · route:inline · tree:a026137ded202e7cbe72816f76907bb7dcb026c2
- [x] T3 · la frase de refused de review dice que correr, y su test la fija   R1 · route:inline · tree:2cd82cf40a0944f86ec0d4b70bcd2cfdc1ec3d8e
- [ ] T4 · el changelog nombra 0292 bajo Unreleased   R1 · route:inline

## History
L1 · "el ratchet mira la negativa que el humano lee, no el constructor que la arma: la de review que hoy no dice que correr pasa a decirlo, y el mecanismo la vigila"
