package akip

//:SDSLSCPI#
//:TIMebase:HOFFset %d
//:TIMebase:SCALe %s
//:CHANnel%2$d:SCALe %1$s
//STARTBIN

func CMD(adr string, cmd string) []byte {
	err := StartClient(adr) // открытие сокета
	if err == nil {
		defer StopClient() // закрытие сокета
		Write(cmd)         // запрос
		//time.Sleep(300 * time.Millisecond) //ожидание ответа
		responce := Read()
		return responce
	}
	return nil
}
