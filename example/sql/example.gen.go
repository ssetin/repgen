// Auto generated code. DO NOT EDIT.
package db

import (
	"database/sql"

	"github.com/ssetin/repgen/example"
)

// UnitRepository =======================================================
type UnitRepository struct {
	tx *sql.Tx
}

func NewUnitRepository(tx example.Transaction) (example.UnitRepository, error) {
	new := &UnitRepository{
		tx: tx.(*sql.Tx),
	}
	return new, nil
}

func (ar UnitRepository) Create(unit *example.Unit) (int, error) {
	var (
		id  int
		row *sql.Row
	)

	if unit.Id == nil {
		row = ar.tx.QueryRow(`INSERT INTO "units"(
	"type", "name", "color", "deleted")
	VALUES ($1, $2, $3, $4) RETURNING "id";`,
			unit.UnitType,
			unit.Name,
			unit.Color,
			unit.Deleted)
	} else {
		row = ar.tx.QueryRow(`INSERT INTO "units"("id", "type", "name", "color", "deleted")
	VALUES ($1, $2, $3, $4, $5) RETURNING "id";`,
			unit.Id,
			unit.UnitType,
			unit.Name,
			unit.Color,
			unit.Deleted)
	}

	err := row.Scan(&id)
	if err != nil {
		return id, err
	}

	err = ar.updateNestedObjects(unit)
	if err != nil {
		return id, err
	}

	return id, nil
}

func (ar UnitRepository) updateNestedObjects(unit *example.Unit) error {
	var err error
	for _, item := range unit.Details {
		_, err = ar.createOrUpdateDetail(item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ar UnitRepository) Update(unit *example.Unit) error {
	_, err := ar.tx.Exec(`UPDATE "units" SET 
	"type"=$1, "name"=$2, "color"=$3, "deleted"=$4
	WHERE "id"=$5;`,
		unit.UnitType,
		unit.Name,
		unit.Color,
		unit.Deleted, unit.Id)

	if err != nil {
		return err
	}

	err = ar.updateNestedObjects(unit)
	if err != nil {
		return err
	}

	return nil
}

func (ar UnitRepository) serializeUnit(row *sql.Row) (*example.Unit, error) {
	res := &example.Unit{}
	err := row.Scan(&res.Id, &res.UnitType, &res.Name, &res.Color, &res.Deleted)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return res, nil
}

func (ar UnitRepository) serializeUnits(rows *sql.Rows) ([]*example.Unit, error) {
	res := make([]*example.Unit, 0, 2)
	for rows.Next() {
		el := &example.Unit{}
		err := rows.Scan(&el.Id, &el.UnitType, &el.Name, &el.Color, &el.Deleted)
		if err != nil {
			return nil, err
		}
		res = append(res, el)
	}
	return res, nil
}

func (ar UnitRepository) FindRawById(id int) (*example.Unit, error) {
	row := ar.tx.QueryRow(`SELECT "id", "type", "name", "color", "deleted"
				FROM "units" WHERE "id"=$1;`, id)
	return ar.serializeUnit(row)
}

func (ar UnitRepository) FindById(id int) (*example.Unit, error) {
	row := ar.tx.QueryRow(`SELECT "id", "type", "name", "color", "deleted"
			FROM "units" WHERE "units"."deleted" is null and "id"=$1;`, id)
	return ar.serializeUnit(row)
}

func (ar UnitRepository) FindByUnitType(unitType string) ([]*example.Unit, error) {
	rows, err := ar.tx.Query(`SELECT "id", "type", "name", "color", "deleted"
			FROM "units" WHERE "units"."deleted" is null and "type"=$1;`, unitType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()
	return ar.serializeUnits(rows)
}

func (ar UnitRepository) MarkDeleted(id int) error {
	_, err := ar.tx.Exec(`UPDATE "units" SET "deleted"=now() WHERE "id"=$1;`, id)
	if err != nil {
		return err
	}
	return nil
}

func (ar UnitRepository) FindByNameAndColor(name string, color string) ([]*example.Unit, error) {
	rows, err := ar.tx.Query(`SELECT "id", "type", "name", "color", "deleted" FROM "units" WHERE
		 "units"."deleted" is null and "name"=$1 and "color"=$2;`, name, color)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()
	return ar.serializeUnits(rows)
}

// Nested Details =======================================================
func (ar UnitRepository) serializeDetail(row *sql.Row) (*example.Detail, error) {
	res := &example.Detail{}
	err := row.Scan(&res.Id, &res.DetailType, &res.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return res, nil
}

func (ar UnitRepository) serializeDetails(rows *sql.Rows) ([]*example.Detail, error) {
	res := make([]*example.Detail, 0, 2)
	for rows.Next() {
		el := &example.Detail{}
		err := rows.Scan(&el.Id, &el.DetailType, &el.Description)
		if err != nil {
			return nil, err
		}
		res = append(res, el)
	}
	return res, nil
}

func (ar UnitRepository) createOrUpdateDetail(detail *example.Detail) (int, error) {
	var (
		id  int
		row *sql.Row
	)
	if detail.Id == nil {
		row = ar.tx.QueryRow(`INSERT INTO "details"(
			"type", "description")
			VALUES ($1, $2) RETURNING "detailId";`,
			detail.DetailType,
			detail.Description)
	} else {
		row = ar.tx.QueryRow(`INSERT INTO "details"("detailId", "type", "description")
			VALUES ($1, $2, $3)
			ON CONFLICT ("detailId") DO UPDATE set
				"type"=EXCLUDED."type",
				"description"=EXCLUDED."description" RETURNING "detailId"`,
			detail.Id,
			detail.DetailType,
			detail.Description)
	}

	err := row.Scan(&id)
	if err != nil {
		return id, err
	}
	return id, nil
}

func (ar UnitRepository) updateDetail(detail *example.Detail) error {
	_, err := ar.tx.Exec(`UPDATE "details" SET 
			"type"=$1, "description"=$2
			WHERE "detailId"=$3;`,
		detail.DetailType,
		detail.Description, detail.Id)

	if err != nil {
		return err
	}
	return nil
}

func (ar UnitRepository) LoadDetails(unit *example.Unit) error {
	rows, err := ar.tx.Query(`SELECT "detailId", "type", "description"
			FROM "details" WHERE "detailId"=$1;`, unit.Id)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	defer rows.Close()
	obj, err := ar.serializeDetails(rows)
	if err != nil {
		return err
	}
	unit.Details = obj
	return nil
}

func (ar UnitRepository) FindByDetailsDetailType(detailType string) ([]*example.Unit, error) {
	rows, err := ar.tx.Query(`SELECT "units"."id", "units"."type", "units"."name", "units"."color", "units"."deleted"
					FROM "units", "details"
					WHERE "units"."id" = "details"."detailId" and
					"units"."deleted" is null and
					
					"details"."type"=$1;`, detailType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()
	return ar.serializeUnits(rows)
}
