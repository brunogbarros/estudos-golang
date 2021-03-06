package servidor

import (
	"crud-basico/banco"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
)

//Letra maiuscula permite acesso as propriedades da Struct, se deixar letra minuscula não é alterado!
type usuario struct {
	ID    int32  `json: "id"`
	Nome  string `json: "nome"`
	Email string `json: "email"`
}

// CriarUsuario - Cria usuários e insere no banco
func CriarUsuario(w http.ResponseWriter, r *http.Request) {
	bodyReq, err := ioutil.ReadAll(r.Body)

	if err != nil {
		//Não farei uso do retornos
		_, _ = w.Write([]byte("Falha ao ler corpo da requisição!"))
		return
	}
	// a requisição vem com um corpo
	var usuario usuario
	if err = json.Unmarshal(bodyReq, &usuario); err != nil {
		_, _ = w.Write([]byte("Erro ao converter usuário!"))
		return
	}
	db, err := banco.Conectar()
	if err != nil {
		_, _ = w.Write([]byte("Erro ao conectar no banco!"))
		return
	}
	defer db.Close()
	// prepare statement
	statement, erro := db.Prepare("insert into usuarios (nome, email) values (?,?)")
	if erro != nil {
		_, _ = w.Write([]byte("Erro ao criar statement"))
		return
	}
	defer statement.Close()

	insert, erro := statement.Exec(usuario.Nome, usuario.Email)
	if erro != nil {
		_, _ = w.Write([]byte("Erro ao executar statement"))
		return
	}
	id, err := insert.LastInsertId()
	if err != nil {
		_, _ = w.Write([]byte("Erro ao retornar Id inserido"))
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("Usuário inserido com sucesso! Id: %d", id)))

}

// BuscarUsuarios -  Traz uma lista de usuários do banco
func BuscarUsuarios(w http.ResponseWriter, r *http.Request) {
	db, erro := banco.Conectar()
	if erro != nil {
		w.Write([]byte("Erro ao conectar no banco!"))
		return
	}
	defer db.Close()

	linhasDoDb, erro := db.Query("select * from usuarios")
	if erro != nil {
		w.Write([]byte("Erro buscar usuários"))
		return
	}
	defer linhasDoDb.Close()

	var usuarios []usuario
	for linhasDoDb.Next() {
		var usuario usuario
		if erro := linhasDoDb.Scan(&usuario.ID, &usuario.Nome, &usuario.Email); erro != nil {
			// Estou ignorando e não tratando o erro que o Write pode retornar!
			w.Write([]byte("Erro ao scannear usuário"))
			return
		}
		// preenche o slice acima
		usuarios = append(usuarios, usuario)
	}
	// retorna o status 200
	w.WriteHeader(http.StatusOK)
	// transformando o slice de usuários em json
	if erro := json.NewEncoder(w).Encode(usuarios); erro != nil {
		w.Write([]byte("Erro converter usuário para JSON"))
		return
	}

}

// BuscarUsuario -  Traz um usuário especifico do banco de dados
func BuscarUsuario(w http.ResponseWriter, r *http.Request) {
	parametros := mux.Vars(r)
	// retorna um map com td recebido
	ID, erro := strconv.ParseUint(parametros["id"], 10, 32)
	if erro != nil {
		w.Write([]byte("Erro ao converter parametro ID "))
		return
	}
	db, erro := banco.Conectar()
	if erro != nil {
		w.Write([]byte("Erro ao conectar no banco"))
		return
	}
	linha, err := db.Query("select * from usuarios where id = ?", ID)
	if err != nil {
		w.Write([]byte("Erro ao buscar usuário do banco."))
		return
	}
	var usuario usuario
	if linha.Next() {
		// popula o struct acima
		if erro := linha.Scan(&usuario.ID, &usuario.Nome, &usuario.Email); erro != nil {
			w.Write([]byte("Erro instanciar usuario"))
			return
		}
	}
	// caso o usuário permaneça com ID zero, significa que nenhum foi encontrado
	if usuario.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		if erro := json.NewEncoder(w).Encode(usuario); erro != nil {
			w.Write([]byte("Erro converter usuário para JSOn"))
			return
		}
	}

}

// AtualizarUsuario -  Traz um usuário especifico do banco de dados e atualiza os dados
func AtualizarUsuario(w http.ResponseWriter, r *http.Request) {
	parametrosRecebidos := mux.Vars(r)
	ID, erro := strconv.ParseUint(parametrosRecebidos["id"], 10, 32)
	if erro != nil {
		w.Write([]byte("Erro ao converter parametro ID para inteiro."))
	}
	bodyReq, erro := ioutil.ReadAll(r.Body)
	if erro != nil {
		w.Write([]byte("Erro ao ler corpo da requisição"))
		return
	}
	var usuario usuario
	// converte o usuario recebido no corpo da requisição para o struct acima!
	if erro := json.Unmarshal(bodyReq, &usuario); erro != nil {
		w.Write([]byte("Erro ao converter usuario."))
		return
	}
	db, erro := banco.Conectar()
	if erro != nil {
		w.Write([]byte("Erro ao conectar no banco"))
		return
	}
	defer db.Close()
	// qlqr coisa que vá fazer inserção no banco precisa ser 'preparada' antes!
	stmt, erro := db.Prepare("update usuarios set nome = ?, email = ? where id = ?")
	if erro != nil {
		w.Write([]byte("Erro criar statement"))
		return
	}
	defer stmt.Close()
	if _, erro := stmt.Exec(usuario.Nome, usuario.Email, ID); erro != nil {
		w.Write([]byte("Erro ao atualizar usuário"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeletarUsuario -  Deleta usuario by id
func DeletarUsuario(w http.ResponseWriter, r *http.Request) {
	parametro := mux.Vars(r)
	ID, erro := strconv.ParseUint(parametro["id"], 10, 32)
	if erro != nil {
		w.Write([]byte("Erro ao converter ID do usuário"))
		return
	}
	db, erro := banco.Conectar()
	if erro != nil {
		w.Write([]byte("Erro ao conectar no banco"))
		return
	}
	defer db.Close()

	stmt, erro := db.Prepare("delete from usuarios where id = ?")
	if erro != nil {
		w.Write([]byte("Erro criar statement"))
		return
	}
	defer stmt.Close()
	_, err := stmt.Exec(ID)
	if err != nil {
		w.Write([]byte("Erro ao apagar usuário do banco!"))
		return
	}
	w.WriteHeader(http.StatusNoContent)

}
