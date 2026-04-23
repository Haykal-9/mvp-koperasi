package handler

import (
	"encoding/gob"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// CartItem stored in session.
type CartItem struct {
	ProductID int
	Jumlah    int
}

func init() {
	gob.Register([]CartItem{})
}

const cartKey = "cart"

func GetCart(c *gin.Context) []CartItem {
	sess := sessions.Default(c)
	v := sess.Get(cartKey)
	if v == nil {
		return []CartItem{}
	}
	items, ok := v.([]CartItem)
	if !ok {
		return []CartItem{}
	}
	return items
}

func SaveCart(c *gin.Context, items []CartItem) {
	sess := sessions.Default(c)
	sess.Set(cartKey, items)
	_ = sess.Save()
}

func ClearCart(c *gin.Context) {
	sess := sessions.Default(c)
	sess.Delete(cartKey)
	_ = sess.Save()
}

// AddOrIncrement adds product to cart, or increments jumlah if already present.
func AddOrIncrement(c *gin.Context, productID, jumlah int) {
	items := GetCart(c)
	for i := range items {
		if items[i].ProductID == productID {
			items[i].Jumlah += jumlah
			SaveCart(c, items)
			return
		}
	}
	items = append(items, CartItem{ProductID: productID, Jumlah: jumlah})
	SaveCart(c, items)
}

// SetQty sets the absolute quantity for a product (removes if <=0).
func SetQty(c *gin.Context, productID, jumlah int) {
	items := GetCart(c)
	out := items[:0]
	for _, it := range items {
		if it.ProductID == productID {
			if jumlah > 0 {
				it.Jumlah = jumlah
				out = append(out, it)
			}
			continue
		}
		out = append(out, it)
	}
	SaveCart(c, out)
}

// RemoveItem deletes a product from cart.
func RemoveItem(c *gin.Context, productID int) {
	items := GetCart(c)
	out := items[:0]
	for _, it := range items {
		if it.ProductID != productID {
			out = append(out, it)
		}
	}
	SaveCart(c, out)
}

// CartCount returns total number of distinct items.
func CartCount(c *gin.Context) int {
	return len(GetCart(c))
}
