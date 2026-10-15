// Single source of wish copy for the egg hatch and the fruit gacha.

// Steal An Egg tiers (real in-game tier names), low → high.
export const EGG_TIERS = [
  { id: 'common', label: 'Common', pts: 1, weight: 50 },
  { id: 'rare', label: 'Rare', pts: 3, weight: 25 },
  { id: 'epic', label: 'Epic', pts: 5, weight: 13 },
  { id: 'legendary', label: 'Legendary', pts: 10, weight: 7 },
  { id: 'mythic', label: 'Mythic', pts: 15, weight: 4 },
  { id: 'secret', label: 'Secret', pts: 25, weight: 1 },
];

export const EGG_WISHES = {
  common: [
    { pet: 'Gà Con Sinh Nhật', wish: 'Chúc Danh ngày nào cũng vui như được báo nghỉ học bất ngờ!' },
    { pet: 'Mèo Ngủ Gật', wish: 'Chúc Danh ngủ đủ giấc, deadline không bao giờ dí sát nút.' },
    { pet: 'Cún Ham Ăn', wish: 'Chúc tiệc sinh nhật nhiều đồ ăn ngon, ăn hoài không mập!' },
  ],
  rare: [
    { pet: 'Cáo Lanh Lợi', wish: 'Chúc Danh học đâu nhớ đó, thi môn nào qua môn đó.' },
    { pet: 'Cú Thông Thái', wish: 'Chúc tuổi 20 mở khoá thêm thật nhiều kỹ năng mới.' },
    { pet: 'Rùa Kiên Nhẫn', wish: 'Chúc Danh chậm mà chắc, việc gì đã bắt đầu thì làm tới nơi tới chốn.' },
  ],
  epic: [
    { pet: 'Trulimero Trulicina', wish: 'Chúc Danh luôn có bạn bè tốt bên cạnh, team nào cũng carry được.' },
    { pet: 'Rồng Con Lửa', wish: 'Chúc lửa đam mê trong Danh cháy mãi không tắt.' },
    { pet: 'Ong Chăm Chỉ', wish: 'Chúc Danh kết nối được thật nhiều mối quan hệ tốt đẹp ở đại học.' },
  ],
  legendary: [
    { pet: 'Brr Brr Patapim', wish: 'Chúc Danh tự tin theo đuổi điều mình thích, ai nói gì kệ họ!' },
    { pet: 'Phượng Hoàng Vàng', wish: 'Chúc Danh vượt qua mọi áp lực học hành và luôn đứng dậy mạnh mẽ hơn.' },
    { pet: 'Sói Bạc', wish: 'Chúc Danh luôn có “đồng đội” sẵn sàng vào server cùng, ngoài đời cũng vậy.' },
  ],
  mythic: [
    { pet: 'Chillin Chilli', wish: 'Chúc Danh phát triển bản thân mỗi ngày: chill mà vẫn lên level đều đều.' },
    { pet: 'King Mammoth', wish: 'Chúc Danh vững vàng như Mammoth trước mọi thử thách tuổi 20.' },
    { pet: 'Bronto', wish: 'Chúc ước mơ của Danh to lớn như Bronto, và từng bước một đều thành hiện thực.' },
  ],
  secret: [
    { pet: 'Cerberus Sinh Nhật', wish: 'SECRET WISH: Chúc Danh hoàn thành mọi ước mơ và nguyện vọng của mình. Tuổi 20 rực rỡ nhất!' },
  ],
};

/** Random wish for a tier, avoiding ones already in `used` (a Set it updates). */
export function eggWish(tierId, used = new Set()) {
  const list = EGG_WISHES[tierId] || EGG_WISHES.common;
  const fresh = list.filter((w) => !used.has(w));
  const pool = fresh.length ? fresh : list;
  const pick = pool[(Math.random() * pool.length) | 0];
  used.add(pick);
  return pick;
}

export const FRUIT_WISHES = {
  'dragon-west': 'Năm 20 tuổi bùng cháy như Dragon. Bay cao, bay xa, không gì cản nổi!',
  kitsune: 'May mắn cấp Mythical: chúc Danh gặp toàn chuyện tốt, roll đâu trúng đó.',
  dough: 'Chúc Danh dẻo dai như Dough: khó khăn nào cũng vượt qua được.',
  venom: 'Chúc Danh “độc” lạ, cá tính và luôn là chính mình.',
  spirit: 'Chúc tinh thần Danh lúc nào cũng phơi phới, năng lượng full thanh.',
  't-rex': 'Chúc Danh mạnh mẽ như T-Rex, đi tới đâu là ăn trọn tới đó (ăn tiệc nha).',
  buddha: 'Chúc Danh bình an, tâm luôn an yên và có nhiều thời gian bên gia đình.',
  light: 'Chúc con đường phía trước của Danh luôn sáng như Light. Nhanh mà chắc.',
  flame: 'Chúc mỗi ngày của Danh đều ấm áp như bữa tiệc sinh nhật hôm nay.',
  ice: 'Chúc Danh luôn bình tĩnh, cool ngầu trước mọi kỳ thi và mọi trận đấu.',
};

export const FRUIT_DEFAULT_WISH = 'Chúc Danh tuổi mới thật nhiều niềm vui và thành công!';

// Gacha weights by rarity (per fruit). Birthday luck: Mythical is boosted.
export const FRUIT_WEIGHTS = { Common: 34, Uncommon: 30, Rare: 22, Legendary: 14, Mythical: 5 };
