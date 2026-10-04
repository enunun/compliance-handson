import { z } from "zod";

/** メモの本文の最大文字数．apps/apiの検証と合わせる． */
export const MAX_NOTE_LENGTH = 1000;

/** メモを作るときの入力． */
export const noteInputSchema = z.object({
  body: z.string().trim().min(1).max(MAX_NOTE_LENGTH),
});

export type NoteInput = z.infer<typeof noteInputSchema>;

/** APIが返すメモ． */
export type Note = {
  id: number;
  ownerId: number;
  body: string;
  createdAt: string;
};
