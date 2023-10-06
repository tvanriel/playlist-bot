package requeststore

import (
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type RequestModel struct {
        gorm.Model

        By string
        GuildID string
        TrackID uint
}

type RequestRepository struct {
        Mysql *gorm.DB
}

type NewRequestRepositoryParams struct {
        fx.In

        MySQL *gorm.DB
}

func NewRequestRepository(p NewRequestRepositoryParams) (*RequestRepository) {
        return &RequestRepository{
                Mysql: p.MySQL,
        }
}


func MigrateRequestRepository(r *RequestRepository) error {
        return r.Mysql.AutoMigrate(&RequestModel{})
}

func (r *RequestRepository) Append(userId string, guildId string, trackId uint) error {
        tx := r.Mysql.Save(&RequestModel{
                By: userId,
                GuildID: guildId,
                TrackID: trackId,
        })
        return tx.Error

}

func(r *RequestRepository) Length(guildId string) (int, error) {
        var length int64
        tx := r.Mysql.Model(&RequestModel{}).Where("guild_id", guildId).Count(&length)
        return int(length), tx.Error
}

func (r *RequestRepository) Pop(guildId string) (uint, error) {
        var request RequestModel
        tx := r.Mysql.Where("guild_id", guildId).Order("created_at ASC").Limit(1).First(&request)
        if tx.Error != nil {
                return 0, tx.Error
        }


        tx = r.Mysql.Limit(1).Delete(&request)
        if tx.Error != nil {
                return 0, tx.Error
        }

        return request.TrackID, nil

}
