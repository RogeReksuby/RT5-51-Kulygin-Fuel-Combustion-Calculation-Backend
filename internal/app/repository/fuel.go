package repository

import (
	"fmt"
	"repback/internal/app/ds"
	"strings"
)

func (r *Repository) GetFuels() ([]ds.Fuel, error) {
	var fuels []ds.Fuel
	// тут запрос SELECT *
	err := r.db.Find(&fuels).Error
	if err != nil {
		return nil, err
	}
	if len(fuels) == 0 {
		return nil, fmt.Errorf("пустой массив")
	}
	return fuels, nil

}

func (r *Repository) GetFuel(id int) (ds.Fuel, error) {
	fuel := ds.Fuel{}
	err := r.db.Where("id = ?", id).First(&fuel).Error
	if err != nil {
		return ds.Fuel{}, err
	}
	return fuel, nil
}

func (r *Repository) GetFuelsByTitle(title string) ([]ds.Fuel, error) {
	var fuels []ds.Fuel
	err := r.db.Where("name ILIKE ?", "%"+title+"%").Find(&fuels).Error
	if err != nil {
		return nil, err
	}
	return fuels, nil
}

// поправь давай емае
func (r *Repository) GetReqFuels() ([]ds.Fuel, error) {
	// имитация получения списка id топлива в заявке
	reqs := []int{2, 4}
	var reqFuels []ds.Fuel
	fuels, err := r.GetFuels()
	if err != nil {
		return nil, err
	}
	for _, id := range reqs {
		for _, fuel := range fuels {
			if fuel.ID == id {
				reqFuels = append(reqFuels, fuel)
			}
		}
	}
	return reqFuels, nil

}

func (r *Repository) GetFuelByTitleOld(title string) ([]ds.Fuel, error) {
	fuels, err := r.GetFuels()
	if err != nil {
		return []ds.Fuel{}, err
	}

	var result []ds.Fuel
	for _, fuel := range fuels {
		if strings.Contains(strings.ToLower(fuel.Title), strings.ToLower(title)) {
			result = append(result, fuel)
		}
	}
	return result, nil
}

func (r *Repository) GetFuelOld(id int) (ds.Fuel, error) {
	fuels, err := r.GetFuels()
	if err != nil {
		return ds.Fuel{}, err
	}

	for _, fuel := range fuels {
		if fuel.ID == id {
			return fuel, nil
		}
	}
	return ds.Fuel{}, fmt.Errorf("отсутсвует топливо")
}

func (r *Repository) GetFuelsOld() ([]ds.Fuel, error) {
	fuels := []ds.Fuel{
		{
			ID:        1,
			Title:     "Метан",
			Heat:      50.1,
			CardImage: "http://127.0.0.1:9000/ripimages/metan.png",
			ShortDesc: "Безопасное, экологичное и экономичное моторное топливо," +
				" получаемое из природного газа",
			FullDesc: "Метан является экологически чистым, экономически выгодным" +
				" и безопасным топливом, получаемым из природного газа. Он" +
				" используется в качестве моторного топлива для транспорта " +
				"(компримированный природный газ - КПГ или CNG), а также для" +
				" отопления и приготовления пищи. Его преимущества включают низкие" +
				" выбросы вредных веществ, долговечность двигателя и низкую стоимость",
		},
		{
			ID:        2,
			Title:     "Пропан-бутан",
			Heat:      43.8,
			CardImage: "http://127.0.0.1:9000/ripimages/propanbutan.jpg",
			ShortDesc: "Универсальное сжиженное углеводородное топливо," +
				" популярное для автономного использования",
			FullDesc: "Пропан-бутановая смесь — это сжиженный углеводородный газ (СУГ)," +
				" широко применяемый в быту, промышленности и качестве автомобильного" +
				" топлива. Благодаря удобству хранения и транспортировки в жидком виде," +
				" он идеален для автономных систем отопления, газовых плит и генераторов." +
				" Основные преимущества: высокая энергоэффективность, стабильность" +
				" поставок и возможность использования в удаленных районах без" +
				" централизованного газоснабжения.",
		},
		{
			ID:        3,
			Title:     "Ацетилен",
			Heat:      50.4,
			CardImage: "http://127.0.0.1:9000/ripimages/ballongaz-acetilen1.png",
			ShortDesc: "Высокоэнергетическое газовое топливо для промышленного применения" +
				" и газопламенной обработки металлов",
			FullDesc: "Ацетилен — это горючий газ с самой высокой температурой горения" +
				" среди углеводородных топлив (до 3150°C). Широко используется в" +
				" промышленности для газовой сварки и резки металлов благодаря" +
				" интенсивному выделению тепла. Применяется в химической промышленности" +
				" для синтеза органических соединений. Требует особых мер безопасности" +
				" из-за высокой взрывоопасности и склонности к самовоспламенению.",
		},
		{
			ID:        4,
			Title:     "Водород",
			Heat:      141,
			CardImage: "http://127.0.0.1:9000/ripimages/hydrogen.jpg",
			ShortDesc: "Перспективное экологичное топливо будущего с нулевыми выбросами," +
				" получаемое из воды и органических источников",
			FullDesc: "Водород — это легкий и энергоемкий газ, считающийся топливом будущего" +
				" благодаря полному отсутствию вредных выбросов при сгорании (образуется" +
				" только вода). Используется в топливных элементах для генерации" +
				" электроэнергии и как чистое ракетное топливо. Производство водорода" +
				" методом электролиза воды позволяет создавать полностью возобновляемую" +
				" энергетическую систему. Основные вызовы: хранение, транспортировка" +
				" и высокая стоимость производства.",
		},
		{
			ID:        5,
			Title:     "Дизельное топливо",
			Heat:      42.7,
			CardImage: "http://127.0.0.1:9000/ripimages/diesel.jpeg",
			ShortDesc: "Высокоэнергетическое жидкое топливо для тяжелой техники," +
				" судов и грузового транспорта",
			FullDesc: "Дизельное топливо — это тяжелая фракция нефти, характеризующаяся" +
				" высокой плотностью энергии и экономичной эффективностью. Широко" +
				" применяется в грузовом транспорте, сельскохозяйственной и строительной" +
				" технике, морских судах и дизельных электростанциях. Отличается" +
				" высоким КПД двигателей и хорошими тяговыми характеристиками." +
				" Современные экологические стандарты требуют использования" +
				" низкосернистых и биодизельных смесей для снижения выбросов.",
		},
	}
	if len(fuels) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}
	return fuels, nil

}
