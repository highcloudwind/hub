import { AlertTriangleIcon, CopyIcon } from "lucide-react";
import React from "react";
import AppHeader from "src/components/AppHeader";
import AppCard from "src/components/connections/AppCard";
import ExternalLink from "src/components/ExternalLink";
import { IsolatedAppTopupDialog } from "src/components/IsolatedAppTopupDialog";
import Loading from "src/components/Loading";
import PasswordInput from "src/components/password/PasswordInput";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "src/components/ui/accordion";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { Textarea } from "src/components/ui/textarea";
import { useToast } from "src/components/ui/use-toast";
import UpgradeCard from "src/components/UpgradeCard";
import { UpgradeDialog } from "src/components/UpgradeDialog";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useApp } from "src/hooks/useApp";
import { useApps } from "src/hooks/useApps";
import { useInfo } from "src/hooks/useInfo";
import { useNodeConnectionInfo } from "src/hooks/useNodeConnectionInfo";
import { copyToClipboard } from "src/lib/clipboard";
import { createApp } from "src/requests/createApp";
import { ConnectAppCard } from "src/screens/apps/AppCreated";
import { CreateAppRequest } from "src/types";
import { handleRequestError } from "src/utils/handleRequestError";

export function UncleJim() {
  const [name, setName] = React.useState("");
  const [appPublicKey, setAppPublicKey] = React.useState("");
  const [connectionSecret, setConnectionSecret] = React.useState("");
  const { data: apps } = useApps();
  const { data: app } = useApp(appPublicKey, true);
  const { data: nodeConnectionInfo } = useNodeConnectionInfo();
  const { toast } = useToast();
  const [isLoading, setLoading] = React.useState(false);
  const { data: info } = useInfo();
  const { data: albyMe, error: albyMeError } = useAlbyMe();

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoading(true);

    try {
      const createAppRequest: CreateAppRequest = {
        name,
        scopes: [
          "get_balance",
          "get_info",
          "list_transactions",
          "lookup_invoice",
          "make_invoice",
          "notifications",
          "pay_invoice",
        ],
        isolated: true,
        metadata: {
          app_store_app_id: "uncle-jim",
        },
      };

      const createAppResponse = await createApp(createAppRequest);

      setConnectionSecret(createAppResponse.pairingUri);
      setAppPublicKey(createAppResponse.pairingPublicKey);

      toast({ title: "New sub-wallet created for " + name });
    } catch (error) {
      handleRequestError(toast, "Failed to create app", error);
    }
    setLoading(false);
  };

  const albyAccountUrl = `https://getalby.com/nwc/new#${connectionSecret}`;
  const valueTag = `<podcast:value type="lightning" method="keysend">
  <podcast:valueRecipient name="${name}" type="node" address="${nodeConnectionInfo?.pubkey}" customKey="696969"  customValue="${app?.id}" split="100"/>
</podcast:value>`;

  const onboardedApps = apps?.filter(
    (app) => app.metadata?.app_store_app_id === "uncle-jim"
  );

  const showForm =
    albyMe?.subscription.plan_code ||
    (onboardedApps && onboardedApps?.length < 3);

  if (!info || (info.albyAccountConnected && !albyMe && !albyMeError)) {
    // make sure to not render the incorrect component
    return <Loading />;
  }

  return (
    <div className="grid gap-5">
      <AppHeader
        title="Friends & Family"
        description="Create sub-wallets for your friends and family powered by your Hub"
        contentRight={
          <UpgradeDialog>
            <Button variant="premium">Upgrade</Button>
          </UpgradeDialog>
        }
      />
      {!connectionSecret && (
        <>
          {showForm ? (
            <form
              onSubmit={handleSubmit}
              className="flex flex-col items-start gap-5 max-w-lg"
            >
              <div className="w-full grid gap-1.5">
                <Label htmlFor="name">Name of friend or family member</Label>
                <Input
                  autoFocus
                  type="text"
                  name="name"
                  value={name}
                  id="name"
                  onChange={(e) => setName(e.target.value)}
                  required
                  autoComplete="off"
                />
              </div>
              <LoadingButton loading={isLoading} type="submit">
                Create Sub-wallet
              </LoadingButton>
            </form>
          ) : (
            <UpgradeCard
              title="Need more Sub-wallets?"
              description="Upgrade to Pro to unlock unlimited sub-wallets"
            />
          )}

          {!!onboardedApps?.length && (
            <>
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 items-stretch app-list">
                {onboardedApps.map((app, index) => (
                  <AppCard key={index} app={app} />
                ))}
              </div>{" "}
            </>
          )}
        </>
      )}
      {connectionSecret && (
        <div className="grid gap-5 max-w-lg">
          <Alert>
            <AlertTriangleIcon className="h-4 w-4" />
            <AlertTitle>Onboard {name} now</AlertTitle>
            <AlertDescription>
              For security reasons these wallet connection details will only be
              shown once and can not be retrieved afterwards. You can save them
              and keep them somewhere safe.
            </AlertDescription>
          </Alert>
          <Accordion type="single" collapsible defaultValue="mobile">
            <AccordionItem value="mobile">
              <AccordionTrigger>Alby Go Mobile App</AccordionTrigger>
              <AccordionContent>
                <p className="text-muted-foreground text-sm mb-5">
                  1. Ask {name} to download the Alby Go app from Google Play or
                  the iOS App Store
                </p>
                <p className="text-muted-foreground text-sm mb-5">
                  2. Ask {name} to scan the below QR code.
                </p>
                {app && (
                  <ConnectAppCard app={app} pairingUri={connectionSecret} />
                )}
              </AccordionContent>
            </AccordionItem>
            <AccordionItem value="account">
              <AccordionTrigger>Alby Account</AccordionTrigger>
              <AccordionContent>
                <p className="text-muted-foreground text-sm mb-5">
                  1. Send {name} an{" "}
                  <ExternalLink
                    to="https://getalby.com/invite_codes"
                    className="underline"
                  >
                    Alby Account invitation
                  </ExternalLink>{" "}
                  if they don't have one yet.
                </p>
                <p className="text-muted-foreground text-sm mb-5">
                  2. Send {name} the below URL which when they open it in their
                  browser, will automatically connect the new wallet to their
                  Alby Account. Do not to share this publicly as it contains the
                  connection secret for their wallet.
                </p>
                <div className="flex gap-2">
                  <PasswordInput disabled readOnly value={albyAccountUrl} />
                  <Button
                    onClick={() => copyToClipboard(albyAccountUrl, toast)}
                    variant="outline"
                  >
                    <CopyIcon className="w-4 h-4 mr-2" />
                    Copy URL
                  </Button>
                </div>
                <Alert className="mt-5">
                  <AlertTriangleIcon className="h-4 w-4" />
                  <AlertTitle>Managing multiple Alby accounts</AlertTitle>
                  <AlertDescription>
                    In case you are managing multiple alby accounts from the
                    same device, we recommend to use multiple browsers (or
                    browser profiles).
                  </AlertDescription>
                </Alert>
              </AccordionContent>
            </AccordionItem>
            <AccordionItem value="extension">
              <AccordionTrigger>Alby Extension</AccordionTrigger>
              <AccordionContent>
                <p className="text-muted-foreground text-sm mb-5">
                  Send {name} the below connection secret which they can add to
                  their Alby Extension by choosing "Bring Your Own Wallet"{" "}
                  {"->"} "Nostr Wallet Connect" and pasting the connection
                  secret. Do not to share this publicly as it contains the
                  connection secret for their wallet.
                </p>
                <div className="flex gap-2">
                  <PasswordInput disabled readOnly value={connectionSecret} />
                  <Button
                    onClick={() => copyToClipboard(connectionSecret, toast)}
                    variant="outline"
                  >
                    <CopyIcon className="w-4 h-4 mr-2" />
                    Copy
                  </Button>
                </div>
              </AccordionContent>
            </AccordionItem>
            <AccordionItem value="other">
              <AccordionTrigger>Other NWC applications</AccordionTrigger>
              <AccordionContent>
                <p className="text-muted-foreground text-sm mb-5">
                  {name} can use any other application that supports NWC. Send
                  the below connection secret. Do not to share this publicly as
                  it contains the connection secret for their wallet.
                </p>
                <div className="flex gap-2">
                  <PasswordInput disabled readOnly value={connectionSecret} />
                  <Button
                    onClick={() => copyToClipboard(connectionSecret, toast)}
                    variant="outline"
                  >
                    <CopyIcon className="w-4 h-4 mr-2" />
                    Copy
                  </Button>
                </div>
              </AccordionContent>
            </AccordionItem>
            <AccordionItem value="podcasting">
              <AccordionTrigger>Podcasting 2.0</AccordionTrigger>
              <AccordionContent>
                <p className="text-muted-foreground text-sm mb-5">
                  To receive podcasting 2.0 payments make sure to give {name}{" "}
                  access to their wallet with one of the options above and share
                  the following details:
                </p>
                <p className="text-muted-foreground text-sm mb-5">
                  <strong>Address:</strong>{" "}
                  <code>{nodeConnectionInfo?.pubkey}</code>
                  <br />
                  <strong>Custom Key:</strong> <code>696969</code>
                  <br />
                  <strong>Custom Value:</strong> <code>{app?.id}</code>
                </p>
                <p>Example podcast:value tag:</p>
                <div className="flex gap-2">
                  <Textarea readOnly className="h-36" value={valueTag} />
                  <Button
                    onClick={() => copyToClipboard(valueTag, toast)}
                    variant="outline"
                  >
                    <CopyIcon className="w-4 h-4 mr-2" />
                    Copy
                  </Button>
                </div>
              </AccordionContent>
            </AccordionItem>
          </Accordion>

          {app && (
            <>
              <p className="mt-5">
                {name} currently has{" "}
                {new Intl.NumberFormat().format(Math.floor(app.balance / 1000))}{" "}
                sats
              </p>
              <IsolatedAppTopupDialog appPubkey={appPublicKey}>
                <Button size="sm" variant="secondary">
                  Increase
                </Button>
              </IsolatedAppTopupDialog>
            </>
          )}
        </div>
      )}
    </div>
  );
}
