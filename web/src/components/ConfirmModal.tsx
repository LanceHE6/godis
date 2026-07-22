import { Modal, ModalContent, ModalHeader, ModalBody, ModalFooter, Button } from '@nextui-org/react'

export default function ConfirmModal({ open, title, message, onConfirm, onCancel }: {
  open: boolean; title: string; message: string; onConfirm: () => void; onCancel: () => void
}) {
  return (
    <Modal isOpen={open} onClose={onCancel} size="sm">
      <ModalContent>
        <ModalHeader>{title}</ModalHeader>
        <ModalBody><p className="text-sm">{message}</p></ModalBody>
        <ModalFooter>
          <Button size="sm" variant="flat" onPress={onCancel}>取消</Button>
          <Button size="sm" color="danger" onPress={onConfirm}>确认删除</Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  )
}
