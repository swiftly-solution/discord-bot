import { AnySelectMenuInteraction, ButtonInteraction, CommandInteraction, Interaction, ModalSubmitInteraction } from "discord.js";
import { anymenus, buttons, localCommands, modals } from "..";

export const type = "on";
export const task = async (interaction: Interaction) => {
    try {
        if (interaction.isButton()) {
            const button = buttons.get(interaction.customId);
            if (!button) return;
            await button(interaction as ButtonInteraction);
        } else if (interaction.isCommand()) {
            const command = localCommands.get(interaction.commandName);
            if (!command) return;
            await command(interaction as CommandInteraction);
        } else if (interaction.isModalSubmit()) {
            const modal = modals.get(interaction.customId);
            if (!modal) return;
            await modal(interaction as ModalSubmitInteraction)
        } else if (interaction.isAnySelectMenu()) {
            const menu = anymenus.get(interaction.customId);
            if (!menu) return;
            await menu(interaction as AnySelectMenuInteraction)
        }
    } catch (err) {
        console.log(err);
        (interaction as CommandInteraction).reply({ content: `An error has been occured, please try again.`, ephemeral: true })
    }
}