package sqlite_test

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
)

func TestDocumentosRepo_ContenidoIdenticoByteAByte(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	documentos := sqlite.NewDocumentosRepo(db)

	inmueble, err := inmuebles.Create(domain.Inmueble{Nombre: "x", Direccion: "y", Tipo: domain.TipoPiso})
	if err != nil {
		t.Fatalf("Create inmueble: %v", err)
	}

	casos := []struct {
		nombre   string
		nombreF  string
		tipoMime string
		contenido []byte
	}{
		{"foto", "fachada.jpg", "image/jpeg", []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F'}},
		{"pdf", "escritura.pdf", "application/pdf", []byte("%PDF-1.4\n...contenido de prueba...")},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			creado, err := documentos.Create(domain.Documento{
				EntidadTipo:   domain.EntidadInmueble,
				EntidadID:     inmueble.ID,
				NombreArchivo: c.nombreF,
				TipoMIME:      c.tipoMime,
				Contenido:     c.contenido,
			})
			if err != nil {
				t.Fatalf("Create documento: %v", err)
			}

			leido, err := documentos.Get(creado.ID)
			if err != nil {
				t.Fatalf("Get documento: %v", err)
			}
			if !bytes.Equal(leido.Contenido, c.contenido) {
				t.Fatalf("el contenido recuperado no es idéntico al original")
			}
			if leido.TamanoBytes != int64(len(c.contenido)) {
				t.Fatalf("tamano_bytes = %d, esperaba %d", leido.TamanoBytes, len(c.contenido))
			}
			if leido.NombreArchivo != c.nombreF || leido.TipoMIME != c.tipoMime {
				t.Fatalf("metadatos no coinciden: %+v", leido)
			}
		})
	}
}

func TestDocumentosRepo_DocumentoDeVariosMB(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	documentos := sqlite.NewDocumentosRepo(db)

	inmueble, err := inmuebles.Create(domain.Inmueble{Nombre: "x", Direccion: "y", Tipo: domain.TipoPiso})
	if err != nil {
		t.Fatalf("Create inmueble: %v", err)
	}

	grande := make([]byte, 6<<20) // 6 MiB
	if _, err := rand.Read(grande); err != nil {
		t.Fatalf("generando contenido de prueba: %v", err)
	}

	creado, err := documentos.Create(domain.Documento{
		EntidadTipo:   domain.EntidadInmueble,
		EntidadID:     inmueble.ID,
		NombreArchivo: "planos.pdf",
		TipoMIME:      "application/pdf",
		Contenido:     grande,
	})
	if err != nil {
		t.Fatalf("Create documento grande: %v", err)
	}

	leido, err := documentos.Get(creado.ID)
	if err != nil {
		t.Fatalf("Get documento grande: %v", err)
	}
	if len(leido.Contenido) != len(grande) {
		t.Fatalf("el documento se truncó: %d bytes leídos, esperaba %d", len(leido.Contenido), len(grande))
	}
	if !bytes.Equal(leido.Contenido, grande) {
		t.Fatalf("el contenido del documento grande no coincide")
	}
}

func TestDocumentosRepo_ListByEntidad(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	documentos := sqlite.NewDocumentosRepo(db)

	a, _ := inmuebles.Create(domain.Inmueble{Nombre: "a", Direccion: "a", Tipo: domain.TipoPiso})
	b, _ := inmuebles.Create(domain.Inmueble{Nombre: "b", Direccion: "b", Tipo: domain.TipoPiso})

	for i := 0; i < 3; i++ {
		if _, err := documentos.Create(domain.Documento{EntidadTipo: domain.EntidadInmueble, EntidadID: a.ID, NombreArchivo: "doc", TipoMIME: "text/plain", Contenido: []byte("x")}); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}
	if _, err := documentos.Create(domain.Documento{EntidadTipo: domain.EntidadInmueble, EntidadID: b.ID, NombreArchivo: "doc", TipoMIME: "text/plain", Contenido: []byte("x")}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	docsA, err := documentos.ListByEntidad(domain.EntidadInmueble, a.ID)
	if err != nil {
		t.Fatalf("ListByEntidad: %v", err)
	}
	if len(docsA) != 3 {
		t.Fatalf("esperaba 3 documentos para el inmueble a, obtuve %d", len(docsA))
	}
	for _, d := range docsA {
		if d.Contenido != nil {
			t.Fatalf("ListByEntidad no debería traer el contenido del BLOB")
		}
	}
}
